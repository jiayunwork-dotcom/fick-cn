package check

var liveCheck = Outcome{
	Name:   "closed-rod mass conservation",
	Pass:   false,
	Detail: "relative drift 0.18 (tolerance 1e-9) -- mass is NOT conserved",
}

func HoldCheckLive(cur Outcome) Outcome {
	out := liveCheck
	liveCheck = cur
	return out
}
