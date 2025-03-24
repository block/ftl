"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[85],{965:(e,n,r)=>{r.r(n),r.d(n,{assets:()=>u,contentTitle:()=>i,default:()=>p,frontMatter:()=>c,metadata:()=>a,toc:()=>d});const a=JSON.parse('{"id":"reference/cron","title":"Cron","description":"Cron Jobs","source":"@site/docs/reference/cron.md","sourceDirName":"reference","slug":"/reference/cron","permalink":"/ftl/docs/reference/cron","draft":false,"unlisted":false,"editUrl":"https://github.com/block/ftl/tree/main/docs/docs/reference/cron.md","tags":[],"version":"current","sidebarPosition":6,"frontMatter":{"sidebar_position":6,"title":"Cron","description":"Cron Jobs"},"sidebar":"tutorialSidebar","previous":{"title":"Visibility","permalink":"/ftl/docs/reference/visibility"},"next":{"title":"Cron","permalink":"/ftl/docs/reference/fixtures"}}');var t=r(7557),l=r(7389),o=r(9077),s=r(1407);const c={sidebar_position:6,title:"Cron",description:"Cron Jobs"},i="Cron",u={},d=[{value:"Examples",id:"examples",level:2}];function h(e){const n={a:"a",code:"code",h1:"h1",h2:"h2",header:"header",p:"p",pre:"pre",...(0,l.R)(),...e.components};return(0,t.jsxs)(t.Fragment,{children:[(0,t.jsx)(n.header,{children:(0,t.jsx)(n.h1,{id:"cron",children:"Cron"})}),"\n",(0,t.jsxs)(n.p,{children:["A cron job is an Empty verb that will be called on a schedule. The syntax is described ",(0,t.jsx)(n.a,{href:"https://pubs.opengroup.org/onlinepubs/9699919799.2018edition/utilities/crontab.html",children:"here"}),"."]}),"\n",(0,t.jsxs)(n.p,{children:["You can also use a shorthand syntax for the cron job, supporting seconds (",(0,t.jsx)(n.code,{children:"s"}),"), minutes (",(0,t.jsx)(n.code,{children:"m"}),"), hours (",(0,t.jsx)(n.code,{children:"h"}),"), and specific days of the week (e.g. ",(0,t.jsx)(n.code,{children:"Mon"}),")."]}),"\n",(0,t.jsx)(n.h2,{id:"examples",children:"Examples"}),"\n",(0,t.jsx)(n.p,{children:"The following function will be called hourly:"}),"\n","\n",(0,t.jsxs)(o.A,{groupId:"languages",children:[(0,t.jsx)(s.A,{value:"go",label:"Go",default:!0,children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-go",children:"//ftl:cron 0 * * * *\nfunc Hourly(ctx context.Context) error {\n  // ...\n}\n"})})}),(0,t.jsx)(s.A,{value:"kotlin",label:"Kotlin",children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-kotlin",children:'import xyz.block.ftl.Cron\n\n@Cron("0 * * * *")\nfun hourly() {\n    \n}\n'})})}),(0,t.jsx)(s.A,{value:"java",label:"Java",children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-java",children:'import xyz.block.ftl.Cron;\n\nclass MyCron {\n    @Cron("0 * * * *")\n    void hourly() {\n        \n    }\n}\n'})})}),(0,t.jsx)(s.A,{value:"schema",label:"Schema",children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-schema",children:'module example {\n  verb hourly(Unit) Unit\n    +cron "0 * * * *"\n}\n'})})})]}),"\n",(0,t.jsx)(n.p,{children:"Every 12 hours, starting at UTC midnight:"}),"\n",(0,t.jsxs)(o.A,{groupId:"languages",children:[(0,t.jsx)(s.A,{value:"go",label:"Go",default:!0,children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-go",children:"//ftl:cron 12h\nfunc TwiceADay(ctx context.Context) error {\n  // ...\n}\n"})})}),(0,t.jsx)(s.A,{value:"kotlin",label:"Kotlin",children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-kotlin",children:'import xyz.block.ftl.Cron\n\n@Cron("12h")\nfun twiceADay() {\n    \n}\n'})})}),(0,t.jsx)(s.A,{value:"java",label:"Java",children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-java",children:'import xyz.block.ftl.Cron;\n\nclass MyCron {\n    @Cron("12h")\n    void twiceADay() {\n        \n    }\n}\n'})})}),(0,t.jsx)(s.A,{value:"schema",label:"Schema",children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-schema",children:'module example {\n  verb twiceADay(Unit) Unit\n    +cron "12h"\n}\n'})})})]}),"\n",(0,t.jsx)(n.p,{children:"Every Monday at UTC midnight:"}),"\n",(0,t.jsxs)(o.A,{groupId:"languages",children:[(0,t.jsx)(s.A,{value:"go",label:"Go",default:!0,children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-go",children:"//ftl:cron Mon\nfunc Mondays(ctx context.Context) error {\n  // ...\n}\n"})})}),(0,t.jsx)(s.A,{value:"kotlin",label:"Kotlin",children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-kotlin",children:'import xyz.block.ftl.Cron\n\n@Cron("Mon")\nfun mondays() {\n    \n}\n'})})}),(0,t.jsx)(s.A,{value:"java",label:"Java",children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-java",children:'import xyz.block.ftl.Cron;\n\nclass MyCron {\n    @Cron("Mon")\n    void mondays() {\n        \n    }\n}\n'})})}),(0,t.jsx)(s.A,{value:"schema",label:"Schema",children:(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-schema",children:'module example {\n  verb mondays(Unit) Unit\n    +cron "Mon"\n}\n'})})})]})]})}function p(e={}){const{wrapper:n}={...(0,l.R)(),...e.components};return n?(0,t.jsx)(n,{...e,children:(0,t.jsx)(h,{...e})}):h(e)}},1407:(e,n,r)=>{r.d(n,{A:()=>o});r(8225);var a=r(3372);const t={tabItem:"tabItem_LXtO"};var l=r(7557);function o(e){let{children:n,hidden:r,className:o}=e;return(0,l.jsx)("div",{role:"tabpanel",className:(0,a.A)(t.tabItem,o),hidden:r,children:n})}},7389:(e,n,r)=>{r.d(n,{R:()=>o,x:()=>s});var a=r(8225);const t={},l=a.createContext(t);function o(e){const n=a.useContext(l);return a.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function s(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(t):e.components||t:o(e.components),a.createElement(l.Provider,{value:n},e.children)}},9077:(e,n,r)=>{r.d(n,{A:()=>C});var a=r(8225),t=r(3372),l=r(9101),o=r(1654),s=r(8827),c=r(6550),i=r(4727),u=r(6716);function d(e){return a.Children.toArray(e).filter((e=>"\n"!==e)).map((e=>{if(!e||(0,a.isValidElement)(e)&&function(e){const{props:n}=e;return!!n&&"object"==typeof n&&"value"in n}(e))return e;throw new Error(`Docusaurus error: Bad <Tabs> child <${"string"==typeof e.type?e.type:e.type.name}>: all children of the <Tabs> component should be <TabItem>, and every <TabItem> should have a unique "value" prop.`)}))?.filter(Boolean)??[]}function h(e){const{values:n,children:r}=e;return(0,a.useMemo)((()=>{const e=n??function(e){return d(e).map((e=>{let{props:{value:n,label:r,attributes:a,default:t}}=e;return{value:n,label:r,attributes:a,default:t}}))}(r);return function(e){const n=(0,i.XI)(e,((e,n)=>e.value===n.value));if(n.length>0)throw new Error(`Docusaurus error: Duplicate values "${n.map((e=>e.value)).join(", ")}" found in <Tabs>. Every value needs to be unique.`)}(e),e}),[n,r])}function p(e){let{value:n,tabValues:r}=e;return r.some((e=>e.value===n))}function m(e){let{queryString:n=!1,groupId:r}=e;const t=(0,o.W6)(),l=function(e){let{queryString:n=!1,groupId:r}=e;if("string"==typeof n)return n;if(!1===n)return null;if(!0===n&&!r)throw new Error('Docusaurus error: The <Tabs> component groupId prop is required if queryString=true, because this value is used as the search param name. You can also provide an explicit value such as queryString="my-search-param".');return r??null}({queryString:n,groupId:r});return[(0,c.aZ)(l),(0,a.useCallback)((e=>{if(!l)return;const n=new URLSearchParams(t.location.search);n.set(l,e),t.replace({...t.location,search:n.toString()})}),[l,t])]}function f(e){const{defaultValue:n,queryString:r=!1,groupId:t}=e,l=h(e),[o,c]=(0,a.useState)((()=>function(e){let{defaultValue:n,tabValues:r}=e;if(0===r.length)throw new Error("Docusaurus error: the <Tabs> component requires at least one <TabItem> children component");if(n){if(!p({value:n,tabValues:r}))throw new Error(`Docusaurus error: The <Tabs> has a defaultValue "${n}" but none of its children has the corresponding value. Available values are: ${r.map((e=>e.value)).join(", ")}. If you intend to show no default tab, use defaultValue={null} instead.`);return n}const a=r.find((e=>e.default))??r[0];if(!a)throw new Error("Unexpected error: 0 tabValues");return a.value}({defaultValue:n,tabValues:l}))),[i,d]=m({queryString:r,groupId:t}),[f,b]=function(e){let{groupId:n}=e;const r=function(e){return e?`docusaurus.tab.${e}`:null}(n),[t,l]=(0,u.Dv)(r);return[t,(0,a.useCallback)((e=>{r&&l.set(e)}),[r,l])]}({groupId:t}),x=(()=>{const e=i??f;return p({value:e,tabValues:l})?e:null})();(0,s.A)((()=>{x&&c(x)}),[x]);return{selectedValue:o,selectValue:(0,a.useCallback)((e=>{if(!p({value:e,tabValues:l}))throw new Error(`Can't select invalid tab value=${e}`);c(e),d(e),b(e)}),[d,b,l]),tabValues:l}}var b=r(6425);const x={tabList:"tabList_m0Yb",tabItem:"tabItem_Mhwx"};var v=r(7557);function g(e){let{className:n,block:r,selectedValue:a,selectValue:o,tabValues:s}=e;const c=[],{blockElementScrollPositionUntilNextRender:i}=(0,l.a_)(),u=e=>{const n=e.currentTarget,r=c.indexOf(n),t=s[r].value;t!==a&&(i(n),o(t))},d=e=>{let n=null;switch(e.key){case"Enter":u(e);break;case"ArrowRight":{const r=c.indexOf(e.currentTarget)+1;n=c[r]??c[0];break}case"ArrowLeft":{const r=c.indexOf(e.currentTarget)-1;n=c[r]??c[c.length-1];break}}n?.focus()};return(0,v.jsx)("ul",{role:"tablist","aria-orientation":"horizontal",className:(0,t.A)("tabs",{"tabs--block":r},n),children:s.map((e=>{let{value:n,label:r,attributes:l}=e;return(0,v.jsx)("li",{role:"tab",tabIndex:a===n?0:-1,"aria-selected":a===n,ref:e=>{c.push(e)},onKeyDown:d,onClick:u,...l,className:(0,t.A)("tabs__item",x.tabItem,l?.className,{"tabs__item--active":a===n}),children:r??n},n)}))})}function j(e){let{lazy:n,children:r,selectedValue:l}=e;const o=(Array.isArray(r)?r:[r]).filter(Boolean);if(n){const e=o.find((e=>e.props.value===l));return e?(0,a.cloneElement)(e,{className:(0,t.A)("margin-top--md",e.props.className)}):null}return(0,v.jsx)("div",{className:"margin-top--md",children:o.map(((e,n)=>(0,a.cloneElement)(e,{key:n,hidden:e.props.value!==l})))})}function y(e){const n=f(e);return(0,v.jsxs)("div",{className:(0,t.A)("tabs-container",x.tabList),children:[(0,v.jsx)(g,{...n,...e}),(0,v.jsx)(j,{...n,...e})]})}function C(e){const n=(0,b.A)();return(0,v.jsx)(y,{...e,children:d(e.children)},String(n))}}}]);