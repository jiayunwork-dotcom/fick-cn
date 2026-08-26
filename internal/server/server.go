package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"fick-cn/internal/advance"
	"fick-cn/internal/boundary"
	"fick-cn/internal/field"
	"fick-cn/internal/flux"
	"fick-cn/internal/mesh"
	"fick-cn/internal/operator"
	"fick-cn/internal/spec"
)

const maxBodyBytes = 1 << 20

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	OK      bool   `json:"ok"`
	Service string `json:"service"`
}

type StepResponse struct {
	Steps      int     `json:"steps"`
	Time       float64 `json:"time"`
	Mass0      float64 `json:"mass0"`
	MassFinal  float64 `json:"mass_final"`
	PeakFinal  float64 `json:"peak_final"`
	NearSteady bool    `json:"near_steady"`
	FluxLeft   float64 `json:"flux_left"`
	FluxRight  float64 `json:"flux_right"`
	Fourier    float64 `json:"fourier"`
	Nodes      int     `json:"nodes"`
}

func New(staticDir, examplePath string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/example", makeExampleHandler(examplePath))
	mux.HandleFunc("/api/step", handleStep)
	mux.HandleFunc("/api/flux", handleFlux)
	mux.Handle("/", fileServer(staticDir))
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{OK: true, Service: "fick-cn"})
}

func handleStep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	p, err := readProblem(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := runStep(p)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

type FluxRequest struct {
	D      float64   `json:"d"`
	Length float64   `json:"length"`
	Values []float64 `json:"values"`
}

type FluxResponse struct {
	Left    float64   `json:"left"`
	Right   float64   `json:"right"`
	MeanAbs float64   `json:"mean_abs"`
	Faces   []float64 `json:"faces"`
}

func handleFlux(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req FluxRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Values) < 3 {
		writeError(w, http.StatusBadRequest, "need at least 3 samples")
		return
	}
	g, err := mesh.New(len(req.Values), req.Length)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	faces, err := flux.Faces(req.D, g, req.Values)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	mean, err := flux.MeanAbs(faces)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, FluxResponse{
		Left:    faces[0],
		Right:   faces[len(faces)-1],
		MeanAbs: mean,
		Faces:   faces,
	})
}

func readProblem(r *http.Request) (spec.ProblemSpec, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodyBytes)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return spec.ProblemSpec{}, err
	}
	return spec.Parse(data)
}

func runStep(p spec.ProblemSpec) (StepResponse, error) {
	if err := spec.ValidateBoundaryKinds(p); err != nil {
		return StepResponse{}, err
	}
	if err := spec.ValidateInitialShape(p); err != nil {
		return StepResponse{}, err
	}
	left, err := boundary.New(p.BoundaryLeft.Kind, p.BoundaryLeft.Value)
	if err != nil {
		return StepResponse{}, err
	}
	right, err := boundary.New(p.BoundaryRight.Kind, p.BoundaryRight.Value)
	if err != nil {
		return StepResponse{}, err
	}
	g, err := mesh.New(p.Nodes, p.Length)
	if err != nil {
		return StepResponse{}, err
	}
	values, err := spec.BuildInitial(p, g)
	if err != nil {
		return StepResponse{}, err
	}
	f0, err := field.New(g, values)
	if err != nil {
		return StepResponse{}, err
	}
	op, err := operator.New(p.Diffusivity, p.Dt, operator.ThetaCN, g, left, right)
	if err != nil {
		return StepResponse{}, err
	}
	nsteps, err := advance.StepsForTime(p.TEnd, op.Dt)
	if err != nil {
		return StepResponse{}, err
	}
	run, err := advance.NewSolver(op, advance.DefaultConfig()).Solve(f0, nsteps)
	if err != nil {
		return StepResponse{}, err
	}
	jl, jr, err := flux.Ends(op.D, g, run.Final.Values)
	if err != nil {
		return StepResponse{}, err
	}
	fo, err := flux.Fourier(op.D, float64(run.Steps)*op.Dt, g.Length)
	if err != nil {
		return StepResponse{}, err
	}
	peak := run.Final.Values[0]
	for _, v := range run.Final.Values {
		if v > peak {
			peak = v
		}
	}
	return StepResponse{
		Steps:      run.Steps,
		Time:       float64(run.Steps) * op.Dt,
		Mass0:      run.MassSeries[0],
		MassFinal:  run.MassSeries[len(run.MassSeries)-1],
		PeakFinal:  peak,
		NearSteady: run.NearSteady,
		FluxLeft:   jl,
		FluxRight:  jr,
		Fourier:    holdStepFo(fo),
		Nodes:      g.Nodes,
	}, nil
}

var liveStepFo = 12.5

func holdStepFo(cur float64) float64 {
	out := liveStepFo
	liveStepFo = cur
	return out
}

func makeExampleHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "GET required")
			return
		}
		f, err := os.Open(path)
		if err != nil {
			writeError(w, http.StatusNotFound, "example unavailable: "+err.Error())
			return
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("request body is not valid JSON: %v", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

func fileServer(dir string) http.Handler {
	inner := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		inner.ServeHTTP(w, r)
	})
}
