"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[364],{129:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>c,contentTitle:()=>i,default:()=>h,frontMatter:()=>o,metadata:()=>r,toc:()=>d});const r=JSON.parse('{"id":"reference/fixtures","title":"Cron","description":"Cron Jobs","source":"@site/docs/reference/fixtures.md","sourceDirName":"reference","slug":"/reference/fixtures","permalink":"/ftl/docs/reference/fixtures","draft":false,"unlisted":false,"editUrl":"https://github.com/block/ftl/tree/main/docs/docs/reference/fixtures.md","tags":[],"version":"current","sidebarPosition":6,"frontMatter":{"sidebar_position":6,"title":"Cron","description":"Cron Jobs"},"sidebar":"tutorialSidebar","previous":{"title":"Cron","permalink":"/ftl/docs/reference/cron"},"next":{"title":"Unit Tests","permalink":"/ftl/docs/reference/unittests"}}');var a=n(7557),s=n(7389),l=n(9077),u=n(1407);const o={sidebar_position:6,title:"Cron",description:"Cron Jobs"},i="Fixtures",c={},d=[{value:"Examples",id:"examples",level:2}];function f(e){const t={code:"code",h1:"h1",h2:"h2",header:"header",p:"p",pre:"pre",...(0,s.R)(),...e.components};return(0,a.jsxs)(a.Fragment,{children:[(0,a.jsx)(t.header,{children:(0,a.jsx)(t.h1,{id:"fixtures",children:"Fixtures"})}),"\n",(0,a.jsx)(t.p,{children:"Fixtures are a way to define a set of data that can be pre-populated for dev mode. Fixtures are essentially a verb that is called when the service is started in dev mode."}),"\n",(0,a.jsx)(t.p,{children:"When not running in dev mode fixtures are not called, and are not present in the schema. This is to prevent fixtures from being called in production. This may change in the future\nto allow fixtures to be used in contract testing between services."}),"\n",(0,a.jsx)(t.p,{children:"FTL also supports manual fixtures that can be called by the user through the console or CLI. This is useful for defining dev mode helper functions that are not usable in production."}),"\n",(0,a.jsx)(t.p,{children:"Note: Test fixtures are not implemented yet, this will be implemented in the future."}),"\n",(0,a.jsx)(t.h2,{id:"examples",children:"Examples"}),"\n","\n",(0,a.jsxs)(l.A,{groupId:"languages",children:[(0,a.jsx)(u.A,{value:"go",label:"Go",default:!0,children:(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-go",children:"//ftl:fixture\nfunc Fixture(ctx context.Context) error {\n  // ...\n}\n//ftl:fixture manual\nfunc ManualFixture(ctx context.Context) error {\n// ...\n}\n"})})}),(0,a.jsx)(u.A,{value:"kotlin",label:"Kotlin",children:(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-kotlin",children:"import xyz.block.ftl.Fixture\n\n@Fixture\nfun fixture() {\n    \n}\n@Fixture(manual=true)\nfun manualFixture() {\n\n}\n"})})}),(0,a.jsx)(u.A,{value:"java",label:"Java",children:(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-java",children:"import xyz.block.ftl.Fixture;\n\nclass MyFixture {\n    @Fixture\n    void fixture() {\n        \n    }\n    @Fixture(manual=true)\n    void manualFixture() {\n\n    }\n}\n"})})}),(0,a.jsx)(u.A,{value:"schema",label:"Schema",children:(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-schema",children:"module example {\n  verb fixture(Unit) Unit\n    +fixture\n  verb manualFixture(Unit) Unit\n    +fixture manual\n}\n"})})})]})]})}function h(e={}){const{wrapper:t}={...(0,s.R)(),...e.components};return t?(0,a.jsx)(t,{...e,children:(0,a.jsx)(f,{...e})}):f(e)}},1407:(e,t,n)=>{n.d(t,{A:()=>l});n(8225);var r=n(3372);const a={tabItem:"tabItem_LXtO"};var s=n(7557);function l(e){let{children:t,hidden:n,className:l}=e;return(0,s.jsx)("div",{role:"tabpanel",className:(0,r.A)(a.tabItem,l),hidden:n,children:t})}},7389:(e,t,n)=>{n.d(t,{R:()=>l,x:()=>u});var r=n(8225);const a={},s=r.createContext(a);function l(e){const t=r.useContext(s);return r.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function u(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:l(e.components),r.createElement(s.Provider,{value:t},e.children)}},9077:(e,t,n)=>{n.d(t,{A:()=>w});var r=n(8225),a=n(3372),s=n(9101),l=n(1654),u=n(8827),o=n(6550),i=n(4727),c=n(6716);function d(e){return r.Children.toArray(e).filter((e=>"\n"!==e)).map((e=>{if(!e||(0,r.isValidElement)(e)&&function(e){const{props:t}=e;return!!t&&"object"==typeof t&&"value"in t}(e))return e;throw new Error(`Docusaurus error: Bad <Tabs> child <${"string"==typeof e.type?e.type:e.type.name}>: all children of the <Tabs> component should be <TabItem>, and every <TabItem> should have a unique "value" prop.`)}))?.filter(Boolean)??[]}function f(e){const{values:t,children:n}=e;return(0,r.useMemo)((()=>{const e=t??function(e){return d(e).map((e=>{let{props:{value:t,label:n,attributes:r,default:a}}=e;return{value:t,label:n,attributes:r,default:a}}))}(n);return function(e){const t=(0,i.XI)(e,((e,t)=>e.value===t.value));if(t.length>0)throw new Error(`Docusaurus error: Duplicate values "${t.map((e=>e.value)).join(", ")}" found in <Tabs>. Every value needs to be unique.`)}(e),e}),[t,n])}function h(e){let{value:t,tabValues:n}=e;return n.some((e=>e.value===t))}function p(e){let{queryString:t=!1,groupId:n}=e;const a=(0,l.W6)(),s=function(e){let{queryString:t=!1,groupId:n}=e;if("string"==typeof t)return t;if(!1===t)return null;if(!0===t&&!n)throw new Error('Docusaurus error: The <Tabs> component groupId prop is required if queryString=true, because this value is used as the search param name. You can also provide an explicit value such as queryString="my-search-param".');return n??null}({queryString:t,groupId:n});return[(0,o.aZ)(s),(0,r.useCallback)((e=>{if(!s)return;const t=new URLSearchParams(a.location.search);t.set(s,e),a.replace({...a.location,search:t.toString()})}),[s,a])]}function m(e){const{defaultValue:t,queryString:n=!1,groupId:a}=e,s=f(e),[l,o]=(0,r.useState)((()=>function(e){let{defaultValue:t,tabValues:n}=e;if(0===n.length)throw new Error("Docusaurus error: the <Tabs> component requires at least one <TabItem> children component");if(t){if(!h({value:t,tabValues:n}))throw new Error(`Docusaurus error: The <Tabs> has a defaultValue "${t}" but none of its children has the corresponding value. Available values are: ${n.map((e=>e.value)).join(", ")}. If you intend to show no default tab, use defaultValue={null} instead.`);return t}const r=n.find((e=>e.default))??n[0];if(!r)throw new Error("Unexpected error: 0 tabValues");return r.value}({defaultValue:t,tabValues:s}))),[i,d]=p({queryString:n,groupId:a}),[m,b]=function(e){let{groupId:t}=e;const n=function(e){return e?`docusaurus.tab.${e}`:null}(t),[a,s]=(0,c.Dv)(n);return[a,(0,r.useCallback)((e=>{n&&s.set(e)}),[n,s])]}({groupId:a}),x=(()=>{const e=i??m;return h({value:e,tabValues:s})?e:null})();(0,u.A)((()=>{x&&o(x)}),[x]);return{selectedValue:l,selectValue:(0,r.useCallback)((e=>{if(!h({value:e,tabValues:s}))throw new Error(`Can't select invalid tab value=${e}`);o(e),d(e),b(e)}),[d,b,s]),tabValues:s}}var b=n(6425);const x={tabList:"tabList_m0Yb",tabItem:"tabItem_Mhwx"};var v=n(7557);function g(e){let{className:t,block:n,selectedValue:r,selectValue:l,tabValues:u}=e;const o=[],{blockElementScrollPositionUntilNextRender:i}=(0,s.a_)(),c=e=>{const t=e.currentTarget,n=o.indexOf(t),a=u[n].value;a!==r&&(i(t),l(a))},d=e=>{let t=null;switch(e.key){case"Enter":c(e);break;case"ArrowRight":{const n=o.indexOf(e.currentTarget)+1;t=o[n]??o[0];break}case"ArrowLeft":{const n=o.indexOf(e.currentTarget)-1;t=o[n]??o[o.length-1];break}}t?.focus()};return(0,v.jsx)("ul",{role:"tablist","aria-orientation":"horizontal",className:(0,a.A)("tabs",{"tabs--block":n},t),children:u.map((e=>{let{value:t,label:n,attributes:s}=e;return(0,v.jsx)("li",{role:"tab",tabIndex:r===t?0:-1,"aria-selected":r===t,ref:e=>{o.push(e)},onKeyDown:d,onClick:c,...s,className:(0,a.A)("tabs__item",x.tabItem,s?.className,{"tabs__item--active":r===t}),children:n??t},t)}))})}function j(e){let{lazy:t,children:n,selectedValue:s}=e;const l=(Array.isArray(n)?n:[n]).filter(Boolean);if(t){const e=l.find((e=>e.props.value===s));return e?(0,r.cloneElement)(e,{className:(0,a.A)("margin-top--md",e.props.className)}):null}return(0,v.jsx)("div",{className:"margin-top--md",children:l.map(((e,t)=>(0,r.cloneElement)(e,{key:t,hidden:e.props.value!==s})))})}function y(e){const t=m(e);return(0,v.jsxs)("div",{className:(0,a.A)("tabs-container",x.tabList),children:[(0,v.jsx)(g,{...t,...e}),(0,v.jsx)(j,{...t,...e})]})}function w(e){const t=(0,b.A)();return(0,v.jsx)(y,{...e,children:d(e.children)},String(t))}}}]);