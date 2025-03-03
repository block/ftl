"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[678],{776:(e,t,n)=>{n.d(t,{A:()=>s});n(8225);var r=n(3372);const a={tabItem:"tabItem_aj6F"};var i=n(7557);function s(e){let{children:t,hidden:n,className:s}=e;return(0,i.jsx)("div",{role:"tabpanel",className:(0,r.A)(a.tabItem,s),hidden:n,children:t})}},3187:(e,t,n)=>{n.d(t,{A:()=>T});var r=n(8225),a=n(3372),i=n(57),s=n(1654),l=n(5739),o=n(7522),c=n(3803),d=n(6273);function u(e){return r.Children.toArray(e).filter((e=>"\n"!==e)).map((e=>{if(!e||(0,r.isValidElement)(e)&&function(e){const{props:t}=e;return!!t&&"object"==typeof t&&"value"in t}(e))return e;throw new Error(`Docusaurus error: Bad <Tabs> child <${"string"==typeof e.type?e.type:e.type.name}>: all children of the <Tabs> component should be <TabItem>, and every <TabItem> should have a unique "value" prop.`)}))?.filter(Boolean)??[]}function h(e){const{values:t,children:n}=e;return(0,r.useMemo)((()=>{const e=t??function(e){return u(e).map((e=>{let{props:{value:t,label:n,attributes:r,default:a}}=e;return{value:t,label:n,attributes:r,default:a}}))}(n);return function(e){const t=(0,c.XI)(e,((e,t)=>e.value===t.value));if(t.length>0)throw new Error(`Docusaurus error: Duplicate values "${t.map((e=>e.value)).join(", ")}" found in <Tabs>. Every value needs to be unique.`)}(e),e}),[t,n])}function p(e){let{value:t,tabValues:n}=e;return n.some((e=>e.value===t))}function f(e){let{queryString:t=!1,groupId:n}=e;const a=(0,s.W6)(),i=function(e){let{queryString:t=!1,groupId:n}=e;if("string"==typeof t)return t;if(!1===t)return null;if(!0===t&&!n)throw new Error('Docusaurus error: The <Tabs> component groupId prop is required if queryString=true, because this value is used as the search param name. You can also provide an explicit value such as queryString="my-search-param".');return n??null}({queryString:t,groupId:n});return[(0,o.aZ)(i),(0,r.useCallback)((e=>{if(!i)return;const t=new URLSearchParams(a.location.search);t.set(i,e),a.replace({...a.location,search:t.toString()})}),[i,a])]}function x(e){const{defaultValue:t,queryString:n=!1,groupId:a}=e,i=h(e),[s,o]=(0,r.useState)((()=>function(e){let{defaultValue:t,tabValues:n}=e;if(0===n.length)throw new Error("Docusaurus error: the <Tabs> component requires at least one <TabItem> children component");if(t){if(!p({value:t,tabValues:n}))throw new Error(`Docusaurus error: The <Tabs> has a defaultValue "${t}" but none of its children has the corresponding value. Available values are: ${n.map((e=>e.value)).join(", ")}. If you intend to show no default tab, use defaultValue={null} instead.`);return t}const r=n.find((e=>e.default))??n[0];if(!r)throw new Error("Unexpected error: 0 tabValues");return r.value}({defaultValue:t,tabValues:i}))),[c,u]=f({queryString:n,groupId:a}),[x,b]=function(e){let{groupId:t}=e;const n=function(e){return e?`docusaurus.tab.${e}`:null}(t),[a,i]=(0,d.Dv)(n);return[a,(0,r.useCallback)((e=>{n&&i.set(e)}),[n,i])]}({groupId:a}),m=(()=>{const e=c??x;return p({value:e,tabValues:i})?e:null})();(0,l.A)((()=>{m&&o(m)}),[m]);return{selectedValue:s,selectValue:(0,r.useCallback)((e=>{if(!p({value:e,tabValues:i}))throw new Error(`Can't select invalid tab value=${e}`);o(e),u(e),b(e)}),[u,b,i]),tabValues:i}}var b=n(7209);const m={tabList:"tabList_lQ3c",tabItem:"tabItem_vLgG"};var j=n(7557);function v(e){let{className:t,block:n,selectedValue:r,selectValue:s,tabValues:l}=e;const o=[],{blockElementScrollPositionUntilNextRender:c}=(0,i.a_)(),d=e=>{const t=e.currentTarget,n=o.indexOf(t),a=l[n].value;a!==r&&(c(t),s(a))},u=e=>{let t=null;switch(e.key){case"Enter":d(e);break;case"ArrowRight":{const n=o.indexOf(e.currentTarget)+1;t=o[n]??o[0];break}case"ArrowLeft":{const n=o.indexOf(e.currentTarget)-1;t=o[n]??o[o.length-1];break}}t?.focus()};return(0,j.jsx)("ul",{role:"tablist","aria-orientation":"horizontal",className:(0,a.A)("tabs",{"tabs--block":n},t),children:l.map((e=>{let{value:t,label:n,attributes:i}=e;return(0,j.jsx)("li",{role:"tab",tabIndex:r===t?0:-1,"aria-selected":r===t,ref:e=>{o.push(e)},onKeyDown:u,onClick:d,...i,className:(0,a.A)("tabs__item",m.tabItem,i?.className,{"tabs__item--active":r===t}),children:n??t},t)}))})}function y(e){let{lazy:t,children:n,selectedValue:i}=e;const s=(Array.isArray(n)?n:[n]).filter(Boolean);if(t){const e=s.find((e=>e.props.value===i));return e?(0,r.cloneElement)(e,{className:(0,a.A)("margin-top--md",e.props.className)}):null}return(0,j.jsx)("div",{className:"margin-top--md",children:s.map(((e,t)=>(0,r.cloneElement)(e,{key:t,hidden:e.props.value!==i})))})}function g(e){const t=x(e);return(0,j.jsxs)("div",{className:(0,a.A)("tabs-container",m.tabList),children:[(0,j.jsx)(v,{...t,...e}),(0,j.jsx)(y,{...t,...e})]})}function T(e){const t=(0,b.A)();return(0,j.jsx)(g,{...e,children:u(e.children)},String(t))}},3908:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>d,contentTitle:()=>c,default:()=>p,frontMatter:()=>o,metadata:()=>r,toc:()=>u});const r=JSON.parse('{"id":"reference/visibility","title":"Visibility","description":"Managing visibility of FTL declarations","source":"@site/docs/reference/visibility.md","sourceDirName":"reference","slug":"/reference/visibility","permalink":"/ftl/docs/reference/visibility","draft":false,"unlisted":false,"editUrl":"https://github.com/block/ftl/tree/main/docs/docs/reference/visibility.md","tags":[],"version":"current","sidebarPosition":5,"frontMatter":{"sidebar_position":5,"title":"Visibility","description":"Managing visibility of FTL declarations"},"sidebar":"tutorialSidebar","previous":{"title":"Types","permalink":"/ftl/docs/reference/types"},"next":{"title":"Cron","permalink":"/ftl/docs/reference/cron"}}');var a=n(7557),i=n(6039),s=n(3187),l=n(776);const o={sidebar_position:5,title:"Visibility",description:"Managing visibility of FTL declarations"},c="Visibility",d={},u=[{value:"Exporting declarations",id:"exporting-declarations",level:2}];function h(e){const t={a:"a",code:"code",h1:"h1",h2:"h2",header:"header",li:"li",ol:"ol",p:"p",pre:"pre",section:"section",sup:"sup",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,i.R)(),...e.components};return(0,a.jsxs)(a.Fragment,{children:[(0,a.jsx)(t.header,{children:(0,a.jsx)(t.h1,{id:"visibility",children:"Visibility"})}),"\n",(0,a.jsx)(t.p,{children:"By default all declarations in FTL are visible only to the module they're declared in. The implicit visibility of types is that of the first verb or other declaration that references it."}),"\n",(0,a.jsx)(t.h2,{id:"exporting-declarations",children:"Exporting declarations"}),"\n",(0,a.jsx)(t.p,{children:"Exporting a declaration makes it accessible to other modules. Some declarations that are entirely local to a module, such as secrets/config, cannot be exported."}),"\n",(0,a.jsx)(t.p,{children:"Types that are transitively referenced by an exported declaration will be automatically exported unless they were already defined but unexported. In this case, an error will be raised and the type must be explicitly exported."}),"\n","\n",(0,a.jsxs)(s.A,{groupId:"languages",children:[(0,a.jsxs)(l.A,{value:"go",label:"Go",default:!0,children:[(0,a.jsx)(t.p,{children:"The following table describes the go directives used to export the corresponding declaration:"}),(0,a.jsxs)(t.table,{children:[(0,a.jsx)(t.thead,{children:(0,a.jsxs)(t.tr,{children:[(0,a.jsx)(t.th,{children:"Symbol"}),(0,a.jsx)(t.th,{children:"Export syntax"})]})}),(0,a.jsxs)(t.tbody,{children:[(0,a.jsxs)(t.tr,{children:[(0,a.jsx)(t.td,{children:"Verb"}),(0,a.jsx)(t.td,{children:(0,a.jsx)(t.code,{children:"//ftl:verb export"})})]}),(0,a.jsxs)(t.tr,{children:[(0,a.jsx)(t.td,{children:"Data"}),(0,a.jsx)(t.td,{children:(0,a.jsx)(t.code,{children:"//ftl:data export"})})]}),(0,a.jsxs)(t.tr,{children:[(0,a.jsx)(t.td,{children:"Enum/Sum type"}),(0,a.jsx)(t.td,{children:(0,a.jsx)(t.code,{children:"//ftl:enum export"})})]}),(0,a.jsxs)(t.tr,{children:[(0,a.jsx)(t.td,{children:"Typealias"}),(0,a.jsx)(t.td,{children:(0,a.jsx)(t.code,{children:"//ftl:typealias export"})})]}),(0,a.jsxs)(t.tr,{children:[(0,a.jsx)(t.td,{children:"Topic"}),(0,a.jsxs)(t.td,{children:[(0,a.jsx)(t.code,{children:"//ftl:export"})," ",(0,a.jsx)(t.sup,{children:(0,a.jsx)(t.a,{href:"#user-content-fn-1",id:"user-content-fnref-1","data-footnote-ref":!0,"aria-describedby":"footnote-label",children:"1"})})]})]})]})]}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-go",children:"//ftl:verb export\nfunc Verb(ctx context.Context, in In) (Out, error)\n\n//ftl:typealias export\ntype UserID string\n"})})]}),(0,a.jsxs)(l.A,{value:"kotlin",label:"Kotlin",children:[(0,a.jsxs)(t.p,{children:["For Kotlin the ",(0,a.jsx)(t.code,{children:"@Export"})," annotation can be used to export a declaration:"]}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-kotlin",children:"@Verb\n@Export\nfun time(): TimeResponse {\n    // ...\n}\n"})})]}),(0,a.jsxs)(l.A,{value:"java",label:"Java",children:[(0,a.jsxs)(t.p,{children:["For Java the ",(0,a.jsx)(t.code,{children:"@Export"})," annotation can be used to export a declaration:"]}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-java",children:"@Verb\n@Export\nTimeResponse time()  {\n    // ...\n}\n"})})]}),(0,a.jsxs)(l.A,{value:"schema",label:"Schema",children:[(0,a.jsxs)(t.p,{children:["In the FTL schema, exported declarations are prefixed with the ",(0,a.jsx)(t.code,{children:"export"})," keyword:"]}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-schema",children:"module example {\n  export data TimeResponse {\n    time Time\n  }\n  \n  export verb time(Unit) example.TimeResponse\n  \n  export topic events example.Event\n  \n  export typealias UserID String\n}\n"})}),(0,a.jsx)(t.p,{children:"Non-exported declarations are visible only within their module:"}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-schema",children:"module example {\n  data InternalConfig {\n    setting String\n  }\n  \n  verb internalProcess(example.InternalConfig) Unit\n}\n"})})]})]}),"\n","\n",(0,a.jsxs)(t.section,{"data-footnotes":!0,className:"footnotes",children:[(0,a.jsx)(t.h2,{className:"sr-only",id:"footnote-label",children:"Footnotes"}),"\n",(0,a.jsxs)(t.ol,{children:["\n",(0,a.jsxs)(t.li,{id:"user-content-fn-1",children:["\n",(0,a.jsxs)(t.p,{children:["By default, topics do not require any annotations as the declaration itself is sufficient. ",(0,a.jsx)(t.a,{href:"#user-content-fnref-1","data-footnote-backref":"","aria-label":"Back to reference 1",className:"data-footnote-backref",children:"\u21a9"})]}),"\n"]}),"\n"]}),"\n"]})]})}function p(e={}){const{wrapper:t}={...(0,i.R)(),...e.components};return t?(0,a.jsx)(t,{...e,children:(0,a.jsx)(h,{...e})}):h(e)}},6039:(e,t,n)=>{n.d(t,{R:()=>s,x:()=>l});var r=n(8225);const a={},i=r.createContext(a);function s(e){const t=r.useContext(i);return r.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function l(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:s(e.components),r.createElement(i.Provider,{value:t},e.children)}}}]);