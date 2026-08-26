"use strict";
const $ = (id) => document.getElementById(id);
function showError(msg) { $("err-text").textContent = msg; $("err-panel").hidden = false; }
function hideError() { $("err-panel").hidden = true; }
function fill(rows) {
  const tb = $("out-table"); tb.innerHTML = "";
  for (const [k,v] of rows) { const tr = document.createElement("tr"); tr.innerHTML = "<td>"+k+"</td><td>"+v+"</td>"; tb.appendChild(tr); }
}
async function loadExample() {
  hideError();
  try {
    const resp = await fetch("/api/example"); const data = await resp.json();
    if (!resp.ok) throw new Error(data.error || "示例失败");
    $("body").value = JSON.stringify(data, null, 2);
    $("hint").textContent = "已加载闭杆算例。";
  } catch (e) { showError(String(e)); }
}
async function runStep() {
  hideError();
  try {
    const resp = await fetch("/api/step", { method:"POST", headers:{"Content-Type":"application/json"}, body: $("body").value });
    const data = await resp.json();
    if (!resp.ok) throw new Error(data.error || "HTTP "+resp.status);
    $("out-panel").hidden = false;
    fill([
      ["步数", String(data.steps)],
      ["时间", data.time.toPrecision(6)],
      ["初质量", data.mass0.toPrecision(6)],
      ["末质量", data.mass_final.toPrecision(6)],
      ["左通量", data.flux_left.toPrecision(5)],
      ["右通量", data.flux_right.toPrecision(5)],
      ["Fourier", data.fourier.toPrecision(5)]
    ]);
  } catch (e) { showError(String(e)); }
}
$("btn-example").addEventListener("click", loadExample);
$("btn-step").addEventListener("click", runStep);
