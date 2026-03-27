package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// ServeSetupPage serves the visual setup wizard.
// If LLM_API_KEY is already set, redirects to the dashboard.
func (h *Handler) ServeSetupPage(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("LLM_API_KEY") != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(setupPageHTML))
}

// handleSetupPost saves LLM/search config to .env and applies it to the
// running process so the server is ready without a restart.
func (h *Handler) handleSetupPost(w http.ResponseWriter, r *http.Request) {
	var body struct {
		LLMApiKey    string `json:"llm_api_key"`
		LLMBaseURL   string `json:"llm_base_url"`
		LLMModelName string `json:"llm_model_name"`
		TavilyApiKey string `json:"tavily_api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.LLMApiKey == "" {
		writeError(w, http.StatusBadRequest, "llm_api_key is required")
		return
	}

	// Build .env content
	lines := []string{
		"# FRACTURE — gerado pelo setup",
		"LLM_API_KEY=" + body.LLMApiKey,
		"LLM_BASE_URL=" + body.LLMBaseURL,
		"LLM_MODEL_NAME=" + body.LLMModelName,
	}
	if body.TavilyApiKey != "" {
		lines = append(lines, "TAVILY_API_KEY="+body.TavilyApiKey)
	}
	content := strings.Join(lines, "\n") + "\n"

	if err := os.WriteFile(".env", []byte(content), 0600); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save config")
		return
	}

	// Apply to current process without restart
	os.Setenv("LLM_API_KEY", body.LLMApiKey)
	os.Setenv("LLM_BASE_URL", body.LLMBaseURL)
	os.Setenv("LLM_MODEL_NAME", body.LLMModelName)
	if body.TavilyApiKey != "" {
		os.Setenv("TAVILY_API_KEY", body.TavilyApiKey)
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

const setupPageHTML = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>FRACTURE — Configuração</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#0f172a;color:#e2e8f0;min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px}
.container{width:100%;max-width:520px}
.logo{text-align:center;margin-bottom:40px}
.logo h1{font-size:32px;font-weight:800;letter-spacing:-1px;color:#fff}
.logo p{color:#64748b;margin-top:8px;font-size:15px}
.card{background:#1e293b;border:1px solid #334155;border-radius:16px;padding:32px}
.step{display:none}.step.active{display:block}
.step-title{font-size:18px;font-weight:700;color:#f1f5f9;margin-bottom:6px}
.step-desc{font-size:14px;color:#64748b;margin-bottom:24px;line-height:1.5}
.provider-grid{display:grid;grid-template-columns:1fr 1fr;gap:12px;margin-bottom:24px}
.provider-btn{background:#0f172a;border:2px solid #334155;border-radius:12px;padding:16px;cursor:pointer;text-align:center;transition:all .2s;color:#e2e8f0}
.provider-btn:hover{border-color:#6366f1;background:#1a1f35}
.provider-btn.selected{border-color:#6366f1;background:#1e1b4b}
.provider-btn .icon{font-size:24px;margin-bottom:8px}
.provider-btn .name{font-weight:600;font-size:14px}
.provider-btn .rec{font-size:11px;color:#6366f1;margin-top:2px}
label{display:block;font-size:13px;font-weight:600;color:#94a3b8;margin-bottom:8px;text-transform:uppercase;letter-spacing:.5px}
input[type=password],input[type=text]{width:100%;background:#0f172a;border:1px solid #334155;border-radius:10px;padding:12px 16px;font-size:15px;color:#e2e8f0;outline:none;transition:border-color .2s;font-family:monospace}
input:focus{border-color:#6366f1}
.input-hint{font-size:12px;color:#475569;margin-top:6px}
.input-hint a{color:#6366f1;text-decoration:none}
.btn-primary{width:100%;background:#6366f1;color:#fff;border:none;border-radius:10px;padding:14px;font-size:16px;font-weight:600;cursor:pointer;margin-top:24px;transition:background .2s}
.btn-primary:hover{background:#4f46e5}
.btn-primary:disabled{background:#334155;cursor:not-allowed}
.btn-secondary{width:100%;background:transparent;color:#64748b;border:none;padding:10px;font-size:14px;cursor:pointer;margin-top:8px;text-decoration:underline}
.progress{display:flex;gap:8px;margin-bottom:28px}
.progress-dot{flex:1;height:3px;background:#334155;border-radius:2px;transition:background .3s}
.progress-dot.done{background:#6366f1}
.loading{text-align:center;padding:40px 0}
.spinner{width:40px;height:40px;border:3px solid #334155;border-top-color:#6366f1;border-radius:50%;animation:spin .8s linear infinite;margin:0 auto 16px}
@keyframes spin{to{transform:rotate(360deg)}}
.success{text-align:center;padding:20px 0}
.success-icon{font-size:48px;margin-bottom:16px}
.success h2{font-size:22px;font-weight:700;margin-bottom:8px}
.success p{color:#64748b;font-size:14px}
.optional-badge{display:inline-block;background:#1e293b;border:1px solid #334155;color:#64748b;font-size:11px;padding:2px 8px;border-radius:20px;margin-left:8px;vertical-align:middle}
</style>
</head>
<body>
<div class="container">
  <div class="logo">
    <h1>⚡ FRACTURE</h1>
    <p>Simulação estratégica com 56 mentes brilhantes</p>
  </div>
  <div class="card">
    <div class="progress">
      <div class="progress-dot done" id="dot1"></div>
      <div class="progress-dot" id="dot2"></div>
      <div class="progress-dot" id="dot3"></div>
    </div>
    <div class="step active" id="step1">
      <div class="step-title">Qual IA você quer usar?</div>
      <div class="step-desc">O FRACTURE usa IA para simular debates estratégicos entre os 56 agentes. Escolha seu provedor.</div>
      <div class="provider-grid">
        <div class="provider-btn selected" id="btn-anthropic" onclick="selectProvider('anthropic')">
          <div class="icon">🧠</div>
          <div class="name">Anthropic</div>
          <div class="rec">Recomendado</div>
        </div>
        <div class="provider-btn" id="btn-openai" onclick="selectProvider('openai')">
          <div class="icon">🤖</div>
          <div class="name">OpenAI</div>
          <div class="rec">GPT-4o</div>
        </div>
      </div>
      <button class="btn-primary" onclick="goStep(2)">Continuar →</button>
    </div>
    <div class="step" id="step2">
      <div class="step-title">Cole sua chave de API</div>
      <div class="step-desc" id="key-desc">Sua chave Anthropic começa com sk-ant-...</div>
      <label>API KEY</label>
      <input type="password" id="api-key" placeholder="sk-ant-api03-..." oninput="validateKey()"/>
      <div class="input-hint" id="key-hint">
        <a href="https://console.anthropic.com" target="_blank">Criar chave no Anthropic Console</a>
      </div>
      <button class="btn-primary" id="btn-step2" onclick="goStep(3)" disabled>Continuar →</button>
    </div>
    <div class="step" id="step3">
      <div class="step-title">Busca web <span class="optional-badge">opcional</span></div>
      <div class="step-desc">Tavily melhora a qualidade das pesquisas de mercado. Sem ele usa DuckDuckGo (gratuito).</div>
      <label>TAVILY API KEY</label>
      <input type="password" id="tavily-key" placeholder="tvly-..."/>
      <div class="input-hint">
        <a href="https://tavily.com" target="_blank">Criar conta grátis — 1.000 buscas/mês</a>
      </div>
      <button class="btn-primary" onclick="submitSetup()">🚀 Iniciar FRACTURE</button>
      <button class="btn-secondary" onclick="submitSetup()">Pular — usar sem busca web</button>
    </div>
    <div class="step" id="step4">
      <div class="loading">
        <div class="spinner"></div>
        <div id="loading-msg">Configurando...</div>
      </div>
    </div>
    <div class="step" id="step5">
      <div class="success">
        <div class="success-icon">✅</div>
        <h2>Tudo pronto!</h2>
        <p>Redirecionando para o dashboard...</p>
      </div>
    </div>
  </div>
</div>
<script>
let selectedProvider='anthropic';
const providers={
  anthropic:{baseUrl:'https://api.anthropic.com',model:'claude-sonnet-4-20250514',desc:'Sua chave Anthropic começa com sk-ant-...',hint:'<a href="https://console.anthropic.com" target="_blank">Criar chave no Anthropic Console</a>'},
  openai:{baseUrl:'https://api.openai.com',model:'gpt-4o',desc:'Sua chave OpenAI começa com sk-...',hint:'<a href="https://platform.openai.com/api-keys" target="_blank">Criar chave no OpenAI Platform</a>'}
};
function selectProvider(p){
  selectedProvider=p;
  document.querySelectorAll('.provider-btn').forEach(b=>b.classList.remove('selected'));
  document.getElementById('btn-'+p).classList.add('selected');
  const cfg=providers[p];
  document.getElementById('key-desc').textContent=cfg.desc;
  document.getElementById('key-hint').innerHTML=cfg.hint;
}
function goStep(n){
  document.querySelectorAll('.step').forEach(s=>s.classList.remove('active'));
  document.getElementById('step'+n).classList.add('active');
  for(let i=1;i<=n;i++){
    const d=document.getElementById('dot'+i);
    if(d)d.classList.add('done');
  }
}
function validateKey(){
  const key=document.getElementById('api-key').value.trim();
  document.getElementById('btn-step2').disabled=key.length<10;
}
async function submitSetup(){
  goStep(4);
  const msgs=['Salvando configuração...','Verificando conexão...','Quase pronto...'];
  let i=0;
  const interval=setInterval(()=>{
    document.getElementById('loading-msg').textContent=msgs[i%msgs.length];
    i++;
  },1200);
  const cfg=providers[selectedProvider];
  const body={
    llm_api_key:document.getElementById('api-key').value.trim(),
    llm_base_url:cfg.baseUrl,
    llm_model_name:cfg.model,
    tavily_api_key:document.getElementById('tavily-key').value.trim()
  };
  try{
    const res=await fetch('/api/v1/setup',{
      method:'POST',
      headers:{'Content-Type':'application/json'},
      body:JSON.stringify(body)
    });
    clearInterval(interval);
    if(res.ok){goStep(5);setTimeout(()=>{window.location.href='/';},2000);}
    else{alert('Erro ao salvar. Tente novamente.');goStep(3);}
  }catch(e){
    clearInterval(interval);
    alert('Erro de conexão.');
    goStep(3);
  }
}
</script>
</body>
</html>`
