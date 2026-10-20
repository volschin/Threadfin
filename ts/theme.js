const themePreference = matchMedia('(prefers-color-scheme: dark)');
let chosenTheme = 'system';
try { chosenTheme = localStorage.getItem('threadfin-theme') || 'system'; } catch {}
if (!['system','light','dark'].includes(chosenTheme)) chosenTheme = 'system';
function applyTheme() {
  document.documentElement.dataset.bsTheme = chosenTheme === 'system' ? (themePreference.matches ? 'dark' : 'light') : chosenTheme;
}
applyTheme();
themePreference.addEventListener('change',applyTheme);
const themeControl = document.createElement('select');
themeControl.className = 'tf-theme-control form-select';
themeControl.setAttribute('aria-label','Color theme');
for (const [value,label] of [['system','System theme'],['light','Light theme'],['dark','Dark theme']]) {
  const option = new Option(label,value); themeControl.add(option);
}
themeControl.value = chosenTheme;
themeControl.addEventListener('change',()=>{ chosenTheme = themeControl.value; try {localStorage.setItem('threadfin-theme',chosenTheme);} catch {} applyTheme(); });
(document.querySelector('.tf-sidebar-header') || document.querySelector('.tf-auth-card header') || document.querySelector('#headline')).append(themeControl);
