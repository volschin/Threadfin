import {build} from 'esbuild';
import {readFile, writeFile, copyFile, mkdir, readdir} from 'node:fs/promises';

// The legacy controllers intentionally share a lexical scope. Bundle that scope
// once per page, with native ESM imports rather than exposing library globals.
const imports = `
import Dialog from 'bootstrap/js/dist/dialog.js';
import Drawer from 'bootstrap/js/dist/drawer.js';
import Combobox from 'bootstrap/js/dist/combobox.js';
import Chips from 'bootstrap/js/dist/chips.js';
import Tooltip from 'bootstrap/js/dist/tooltip.js';
import ClipboardJS from 'clipboard';
`;
const app = ['network','mapping_state','filter','menu','users_page','mapping','sources','app_state','overview','connections','activity','navigation','settings','settings_page','logs','log_page','base'];
const setup = ['configuration','network','menu','sources','settings','base'];
const theme = await readFile('ts/theme.js', 'utf8');
for (const [name, controllers, start] of [
  ['app', app, 'PageReady();'],
  ['setup', setup, 'document.getElementById("next").addEventListener("click",saveWizard); readyForConfiguration(0);'],
  ['auth', ['authentication'], 'document.getElementById("authentication").addEventListener("submit",event => { if (!login(event)) event.preventDefault(); });'],
]) {
  const source = (await Promise.all(controllers.map(x => readFile(`html/js/${x}_ts.js`,'utf8')))).join('\n');
  await build({stdin:{contents: (name === "auth" ? "" : imports) + theme + '\n' + source + '\n' + start,resolveDir:process.cwd()},bundle:true,format:'esm',target:['chrome130','firefox132','safari18'],outfile:`html/js/${name}.js`,minify:false,legalComments:'eof'});
}
await mkdir('html/css/vendor',{recursive:true});
await copyFile('node_modules/bootstrap/dist/css/bootstrap.min.css','html/css/vendor/bootstrap.min.css');
await copyFile('node_modules/@fortawesome/fontawesome-free/css/all.min.css','html/css/vendor/fontawesome.min.css');
// Font Awesome's stylesheet resolves fonts relative to css/vendor/.
await mkdir('html/css/webfonts',{recursive:true});
for (const file of await readdir('node_modules/@fortawesome/fontawesome-free/webfonts')) {
  if (file.endsWith('.woff2')) await copyFile(`node_modules/@fortawesome/fontawesome-free/webfonts/${file}`,`html/css/webfonts/${file}`);
}
const layers = '@layer colors, config, root, reboot, layout, content, forms, components, custom, helpers, utilities;\n';
const legacy = await Promise.all(['screen','base','app-shell'].map(x=>readFile(`html/css/${x}.css`,'utf8')));
const tokens = await readFile('html/css/modern.css','utf8');
const componentTokens = await readFile('html/css/theme.css','utf8');
await writeFile('html/css/app.css',layers + '@layer custom {\n' + legacy.join('\n').replace(/[ \t]+$/gm, '') + '\n' + tokens + '\n}\n' + componentTokens);
