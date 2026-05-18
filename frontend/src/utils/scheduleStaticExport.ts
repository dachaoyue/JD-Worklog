import type { ScheduleDay, ScheduleTask } from '../api/work_schedule'

export type StaticSchedulePerson = {
  user_id: number
  person_name: string
  tech_direction: string
  tasks: ScheduleTask[]
}

export type StaticSchedulePayload = {
  title: string
  start: string
  end: string
  generatedAt: string
  days: ScheduleDay[]
  persons: StaticSchedulePerson[]
}

const STATIC_CSS = `
*{box-sizing:border-box}
body{margin:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;font-size:12px;color:#303133;background:#f5f7fa}
.wrap{max-width:100%;padding:16px 20px 32px}
h1{margin:0 0 8px;font-size:18px;font-weight:600}
.meta{margin:0 0 16px;color:#606266;font-size:13px;line-height:1.6}
.notice{margin:16px 0 0;padding:10px 12px;background:#ecf5ff;border-radius:4px;color:#409eff;font-size:12px}
.gantt-split{display:flex;border:1px solid #ebeef5;border-radius:4px;background:#fff;overflow:hidden;max-width:100%}
.gantt-fixed{flex-shrink:0;background:#fff;z-index:3;box-shadow:4px 0 8px -4px rgba(0,0,0,.12)}
.gantt-scroll{flex:1;overflow-x:auto;overflow-y:hidden;min-width:0}
.gantt-table{border-collapse:collapse;font-size:12px;table-layout:fixed}
.gantt-table th,.gantt-table td{border:1px solid #ebeef5;text-align:center;vertical-align:middle;white-space:nowrap;height:40px}
.gantt-table thead th{background:#f5f7fa;font-weight:600;padding:6px 4px}
.col-person{width:100px;min-width:100px;padding:6px 8px;text-align:left}
.col-tech{width:80px;min-width:80px;padding:4px 6px;text-align:left}
.person-name{font-weight:500}
.tech-text{color:#606266}
.date-head,.date-cell{width:37px;min-width:37px;padding:0}
.date-head.holiday,.date-cell.holiday{background:#f4f4f5}
.date-cell{position:relative;overflow:visible}
.bar-overlay{position:absolute;left:0;top:0;height:100%;z-index:2;display:flex;align-items:center;box-sizing:border-box;padding:0 4px;font-size:11px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.empty-row{padding:16px;color:#909399;text-align:center}
`

const RENDER_SCRIPT = `(function(){
var DATA=__DATA__;
var CELL_W=37;
function esc(s){if(s==null)return'';return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');}
function formatHeadDate(s){var p=s.split('-');return p.length===3?Number(p[1])+'/'+Number(p[2]):s;}
function textColor(bg){var hex=(bg||'').replace('#','');if(hex.length!==6)return '#fff';var r=parseInt(hex.slice(0,2),16),g=parseInt(hex.slice(2,4),16),b=parseInt(hex.slice(4,6),16);var lum=(0.299*r+0.587*g+0.114*b)/255;return lum>0.62?'#303133':'#ffffff';}
function taskColor(t){return t.color||'#409eff';}
function taskAtDate(person,date){for(var i=0;i<person.tasks.length;i++){var t=person.tasks[i];if(date>=t.start_date&&date<=t.end_date)return t;}return null;}
function isBarStart(task,date){return date===task.start_date;}
function barSpan(task,startIdx){var endIdx=startIdx;for(var i=0;i<DATA.days.length;i++){if(DATA.days[i].date===task.end_date){endIdx=i;break;}}var span=Math.max(1,endIdx-startIdx+1);return{width:'calc('+span+' * '+CELL_W+'px - 1px)'};}
function cellHasSchedule(person,date){return !!taskAtDate(person,date);}
var root=document.getElementById('gantt-root');
if(!root)return;
var h='<div class="gantt-split">';
h+='<div class="gantt-fixed"><table class="gantt-table"><thead><tr><th class="col-person">\u4eba\u5458</th><th class="col-tech">\u6280\u672f\u65b9\u5411</th></tr></thead><tbody>';
if(!DATA.persons.length){h+='<tr><td colspan="2" class="empty-row">\u6682\u65e0\u6392\u671f\u6570\u636e</td></tr>';}
else{for(var pi=0;pi<DATA.persons.length;pi++){var p=DATA.persons[pi];h+='<tr><td class="col-person"><span class="person-name">'+esc(p.person_name)+'</span></td><td class="col-tech"><span class="tech-text">'+esc(p.tech_direction||'')+'</span></td></tr>';}}
h+='</tbody></table></div>';
h+='<div class="gantt-scroll"><table class="gantt-table"><thead><tr>';
for(var di=0;di<DATA.days.length;di++){var day=DATA.days[di];h+='<th class="date-head'+(day.is_holiday?' holiday':'')+'">'+formatHeadDate(day.date)+'</th>';}
h+='</tr></thead><tbody>';
if(!DATA.persons.length){h+='<tr><td colspan="'+DATA.days.length+'" class="empty-row">\u6682\u65e0\u6392\u671f</td></tr>';}
else{for(pi=0;pi<DATA.persons.length;pi++){p=DATA.persons[pi];h+='<tr>';for(di=0;di<DATA.days.length;di++){day=DATA.days[di];var cls='date-cell';if(day.is_holiday&&!cellHasSchedule(p,day.date))cls+=' holiday';h+='<td class="'+cls+'">';for(var ti=0;ti<p.tasks.length;ti++){var task=p.tasks[ti];if(!isBarStart(task,day.date))continue;var bg=taskColor(task);var sp=barSpan(task,di);h+='<div class="bar-overlay" style="width:'+sp.width+';background:'+bg+';color:'+textColor(bg)+'">'+esc(task.label||'')+'</div>';}h+='</td>';}h+='</tr>';}}
h+='</tbody></table></div></div>';
root.innerHTML=h;
})();`

function escapeScriptJson(obj: unknown): string {
  return JSON.stringify(obj).replace(/</g, '\u003c')
}

export function buildScheduleStaticHtml(payload: StaticSchedulePayload): string {
  const dataJson = escapeScriptJson(payload)
  const script = RENDER_SCRIPT.replace('__DATA__', dataJson)
  const title = payload.title || '\u5de5\u4f5c\u6392\u671f'
  const notice =
    '\u672c\u9875\u9762\u4e3a\u5feb\u7167\uff0c\u4ec5\u7528\u4e8e\u67e5\u770b\uff1b\u53ef\u901a\u8fc7\u90ae\u4ef6\u3001\u7f51\u76d8\u6216 U \u76d8\u5206\u4eab\uff0c\u65e0\u9700\u8bbf\u95ee\u670d\u52a1\u5668\u3002'
  return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>${title}</title>
<style>${STATIC_CSS}</style>
</head>
<body>
<div class="wrap">
<h1>${title}</h1>
<p class="meta">
\u5468\u671f\uff1a${payload.start} ~ ${payload.end}<br>
\u751f\u6210\u65f6\u95f4\uff1a${payload.generatedAt}
</p>
<div id="gantt-root"></div>
<p class="notice">${notice}</p>
</div>
<script>${script}</script>
</body>
</html>`
}

export function downloadScheduleStaticPage(payload: StaticSchedulePayload): void {
  const html = buildScheduleStaticHtml(payload)
  const blob = new Blob([html], { type: 'text/html;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  const safeStart = payload.start.replace(/\//g, '-')
  const safeEnd = payload.end.replace(/\//g, '-')
  a.href = url
  a.download = `\u5de5\u4f5c\u6392\u671f_${safeStart}_${safeEnd}.html`
  a.click()
  URL.revokeObjectURL(url)
}
