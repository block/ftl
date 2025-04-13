"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[828],{361:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>u,contentTitle:()=>c,default:()=>m,frontMatter:()=>i,metadata:()=>r,toc:()=>d});const r=JSON.parse('{"id":"reference/verbs","title":"Verbs","description":"Declaring and calling Verbs","source":"@site/docs/reference/verbs.md","sourceDirName":"reference","slug":"/reference/verbs","permalink":"/ftl/docs/reference/verbs","draft":false,"unlisted":false,"editUrl":"https://github.com/block/ftl/tree/main/docs/docs/reference/verbs.md","tags":[],"version":"current","sidebarPosition":1,"frontMatter":{"sidebar_position":1,"title":"Verbs","description":"Declaring and calling Verbs"},"sidebar":"tutorialSidebar","previous":{"title":"Reference","permalink":"/ftl/docs/category/reference"},"next":{"title":"Types","permalink":"/ftl/docs/reference/types"}}');var a=t(7968),l=t(4205),s=t(5624),o=t(7178);const i={sidebar_position:1,title:"Verbs",description:"Declaring and calling Verbs"},c="Verbs",u={},d=[{value:"Defining Verbs",id:"defining-verbs",level:2},{value:"Calling Verbs",id:"calling-verbs",level:2}];function h(e){const n={a:"a",code:"code",h1:"h1",h2:"h2",header:"header",p:"p",pre:"pre",...(0,l.R)(),...e.components};return(0,a.jsxs)(a.Fragment,{children:[(0,a.jsx)(n.header,{children:(0,a.jsx)(n.h1,{id:"verbs",children:"Verbs"})}),"\n",(0,a.jsx)(n.h2,{id:"defining-verbs",children:"Defining Verbs"}),"\n","\n",(0,a.jsxs)(s.A,{groupId:"languages",children:[(0,a.jsxs)(o.A,{value:"go",label:"Go",default:!0,children:[(0,a.jsxs)(n.p,{children:["To declare a Verb, write a normal Go function with the following signature, annotated with the Go ",(0,a.jsx)(n.a,{href:"https://tip.golang.org/doc/comment#syntax",children:"comment directive"})," ",(0,a.jsx)(n.code,{children:"//ftl:verb"}),":"]}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-go",children:"//ftl:verb\nfunc F(context.Context, In) (Out, error) { }\n"})}),(0,a.jsx)(n.p,{children:"eg."}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-go",children:"type EchoRequest struct {}\n\ntype EchoResponse struct {}\n\n//ftl:verb\nfunc Echo(ctx context.Context, in EchoRequest) (EchoResponse, error) {\n  // ...\n}\n"})})]}),(0,a.jsxs)(o.A,{value:"kotlin",label:"Kotlin",children:[(0,a.jsxs)(n.p,{children:["To declare a Verb, write a normal Kotlin function with the following signature, annotated with the Kotlin ",(0,a.jsx)(n.a,{href:"https://kotlinlang.org/docs/annotations.html",children:"annotation"})," ",(0,a.jsx)(n.code,{children:"@Verb"}),":"]}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-kotlin",children:"@Verb\nfun F(In): Out { }\n"})}),(0,a.jsx)(n.p,{children:"eg."}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-kotlin",children:"data class EchoRequest\ndata class EchoResponse\n\n@Verb\nfun echo(request: EchoRequest): EchoResponse {\n  // ...\n}\n"})})]}),(0,a.jsxs)(o.A,{value:"java",label:"Java",children:[(0,a.jsxs)(n.p,{children:["To declare a Verb, write a normal Java method with the following signature, annotated with the ",(0,a.jsx)(n.code,{children:"@Verb"})," annotation:"]}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-java",children:"@Verb\npublic Output f(Input input) { }\n"})}),(0,a.jsx)(n.p,{children:"eg."}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-java",children:"import xyz.block.ftl.Verb;\n\nclass EchoRequest {}\n\nclass EchoResponse {}\n\npublic class EchoClass {\n    @Verb\n    public EchoResponse echo(EchoRequest request) {\n        // ...\n    }\n}\n"})})]}),(0,a.jsxs)(o.A,{value:"schema",label:"Schema",children:[(0,a.jsx)(n.p,{children:"In the FTL schema, verbs are declared with their input and output types:"}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-schema",children:"module example {\n  data EchoRequest {}\n  \n  data EchoResponse {}\n  \n  verb echo(example.EchoRequest) example.EchoResponse\n}\n"})}),(0,a.jsx)(n.p,{children:"Verbs can be exported to make them callable from other modules:"}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-schema",children:"module example {\n  export verb echo(example.EchoRequest) example.EchoResponse\n}\n"})})]})]}),"\n",(0,a.jsxs)(n.p,{children:["By default verbs are only visible to other verbs in the same module (see ",(0,a.jsx)(n.a,{href:"./visibility",children:"visibility"})," for more information)."]}),"\n",(0,a.jsx)(n.h2,{id:"calling-verbs",children:"Calling Verbs"}),"\n",(0,a.jsxs)(s.A,{groupId:"languages",children:[(0,a.jsxs)(o.A,{value:"go",label:"Go",default:!0,children:[(0,a.jsxs)(n.p,{children:["To call a verb, import the module's verb client (",(0,a.jsx)(n.code,{children:"{ModuleName}.{VerbName}Client"}),"), add it to your verb's signature, then invoke it as a function. eg."]}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-go",children:"//ftl:verb\nfunc Echo(ctx context.Context, in EchoRequest, tc time.TimeClient) (EchoResponse, error) {\n    out, err := tc(ctx, TimeRequest{...})\n}\n"})}),(0,a.jsxs)(n.p,{children:["Verb clients are generated by FTL. If the callee verb belongs to the same module as the caller, you must build the\nmodule first (with callee verb defined) in order to generate its client for use by the caller. Local verb clients are\navailable in the generated ",(0,a.jsx)(n.code,{children:"types.ftl.go"})," file as ",(0,a.jsx)(n.code,{children:"{VerbName}Client"}),"."]})]}),(0,a.jsxs)(o.A,{value:"kotlin",label:"Kotlin",children:[(0,a.jsxs)(n.p,{children:["To call a verb, import the module's verb client, add it to your verb's signature, then ",(0,a.jsx)(n.code,{children:"call()"})," it. eg."]}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-kotlin",children:"import ftl.time.TimeClient\nimport xyz.block.ftl.Verb\n\n@Verb\nfun echo(req: EchoRequest, time: TimeClient): EchoResponse {\n  val response = time.call()\n  // ...\n}\n\nval response = time.call()\n"})}),(0,a.jsx)(n.p,{children:"Verb clients are generated by FTL. If the callee verb belongs to the same module as the caller, you must manually define your\nown client:"}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-kotlin",children:"interface TimeClient {\n    @VerbClient\n    fun time(): TimeResponse\n}\n"})})]}),(0,a.jsxs)(o.A,{value:"java",label:"Java",children:[(0,a.jsx)(n.p,{children:"To call a verb, import the module's verb client, add it to your verb's signature, then call it. eg."}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-java",children:"import ftl.time.TimeClient;\nimport xyz.block.ftl.Verb;\n\npublic class EchoClass {\n    @Verb\n    public EchoResponse echo(EchoRequest request, TimeClient time) {\n        TimeResponse response = time.call();\n        // ...\n    }\n}\n"})}),(0,a.jsx)(n.p,{children:"Verb clients are generated by FTL. If the callee verb belongs to the same module as the caller, you must manually define your\nown client:"}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-java",children:"public interface TimeClient {\n    @VerbClient\n    TimeResponse time();\n}\n"})})]}),(0,a.jsxs)(o.A,{value:"schema",label:"Schema",children:[(0,a.jsxs)(n.p,{children:["In the FTL schema, verb calls are represented by the ",(0,a.jsx)(n.code,{children:"+calls"})," annotation:"]}),(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-schema",children:"module echo {\n  data EchoRequest {}\n  \n  data EchoResponse {}\n  \n  verb echo(example.EchoRequest) example.EchoResponse\n    +calls time.time\n}\n\nmodule time {\n  data TimeRequest {}\n  \n  data TimeResponse {\n    time Time\n  }\n  \n  export verb time(time.TimeRequest) time.TimeResponse\n}\n"})}),(0,a.jsxs)(n.p,{children:["The ",(0,a.jsx)(n.code,{children:"+calls"})," annotation indicates that the verb calls another verb, in this case the ",(0,a.jsx)(n.code,{children:"time"})," verb from the ",(0,a.jsx)(n.code,{children:"time"})," module."]})]})]})]})}function m(e={}){const{wrapper:n}={...(0,l.R)(),...e.components};return n?(0,a.jsx)(n,{...e,children:(0,a.jsx)(h,{...e})}):h(e)}},4205:(e,n,t)=>{t.d(n,{R:()=>s,x:()=>o});var r=t(4700);const a={},l=r.createContext(a);function s(e){const n=r.useContext(l);return r.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function o(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:s(e.components),r.createElement(l.Provider,{value:n},e.children)}},5624:(e,n,t)=>{t.d(n,{A:()=>V});var r=t(4700),a=t(3372),l=t(6128),s=t(3263),o=t(7974),i=t(6605),c=t(1394),u=t(9829);function d(e){return r.Children.toArray(e).filter((e=>"\n"!==e)).map((e=>{if(!e||(0,r.isValidElement)(e)&&function(e){const{props:n}=e;return!!n&&"object"==typeof n&&"value"in n}(e))return e;throw new Error(`Docusaurus error: Bad <Tabs> child <${"string"==typeof e.type?e.type:e.type.name}>: all children of the <Tabs> component should be <TabItem>, and every <TabItem> should have a unique "value" prop.`)}))?.filter(Boolean)??[]}function h(e){const{values:n,children:t}=e;return(0,r.useMemo)((()=>{const e=n??function(e){return d(e).map((e=>{let{props:{value:n,label:t,attributes:r,default:a}}=e;return{value:n,label:t,attributes:r,default:a}}))}(t);return function(e){const n=(0,c.XI)(e,((e,n)=>e.value===n.value));if(n.length>0)throw new Error(`Docusaurus error: Duplicate values "${n.map((e=>e.value)).join(", ")}" found in <Tabs>. Every value needs to be unique.`)}(e),e}),[n,t])}function m(e){let{value:n,tabValues:t}=e;return t.some((e=>e.value===n))}function b(e){let{queryString:n=!1,groupId:t}=e;const a=(0,s.W6)(),l=function(e){let{queryString:n=!1,groupId:t}=e;if("string"==typeof n)return n;if(!1===n)return null;if(!0===n&&!t)throw new Error('Docusaurus error: The <Tabs> component groupId prop is required if queryString=true, because this value is used as the search param name. You can also provide an explicit value such as queryString="my-search-param".');return t??null}({queryString:n,groupId:t});return[(0,i.aZ)(l),(0,r.useCallback)((e=>{if(!l)return;const n=new URLSearchParams(a.location.search);n.set(l,e),a.replace({...a.location,search:n.toString()})}),[l,a])]}function p(e){const{defaultValue:n,queryString:t=!1,groupId:a}=e,l=h(e),[s,i]=(0,r.useState)((()=>function(e){let{defaultValue:n,tabValues:t}=e;if(0===t.length)throw new Error("Docusaurus error: the <Tabs> component requires at least one <TabItem> children component");if(n){if(!m({value:n,tabValues:t}))throw new Error(`Docusaurus error: The <Tabs> has a defaultValue "${n}" but none of its children has the corresponding value. Available values are: ${t.map((e=>e.value)).join(", ")}. If you intend to show no default tab, use defaultValue={null} instead.`);return n}const r=t.find((e=>e.default))??t[0];if(!r)throw new Error("Unexpected error: 0 tabValues");return r.value}({defaultValue:n,tabValues:l}))),[c,d]=b({queryString:t,groupId:a}),[p,f]=function(e){let{groupId:n}=e;const t=function(e){return e?`docusaurus.tab.${e}`:null}(n),[a,l]=(0,u.Dv)(t);return[a,(0,r.useCallback)((e=>{t&&l.set(e)}),[t,l])]}({groupId:a}),x=(()=>{const e=c??p;return m({value:e,tabValues:l})?e:null})();(0,o.A)((()=>{x&&i(x)}),[x]);return{selectedValue:s,selectValue:(0,r.useCallback)((e=>{if(!m({value:e,tabValues:l}))throw new Error(`Can't select invalid tab value=${e}`);i(e),d(e),f(e)}),[d,f,l]),tabValues:l}}var f=t(7582);const x={tabList:"tabList_HKUW",tabItem:"tabItem_l9_f"};var g=t(7968);function v(e){let{className:n,block:t,selectedValue:r,selectValue:s,tabValues:o}=e;const i=[],{blockElementScrollPositionUntilNextRender:c}=(0,l.a_)(),u=e=>{const n=e.currentTarget,t=i.indexOf(n),a=o[t].value;a!==r&&(c(n),s(a))},d=e=>{let n=null;switch(e.key){case"Enter":u(e);break;case"ArrowRight":{const t=i.indexOf(e.currentTarget)+1;n=i[t]??i[0];break}case"ArrowLeft":{const t=i.indexOf(e.currentTarget)-1;n=i[t]??i[i.length-1];break}}n?.focus()};return(0,g.jsx)("ul",{role:"tablist","aria-orientation":"horizontal",className:(0,a.A)("tabs",{"tabs--block":t},n),children:o.map((e=>{let{value:n,label:t,attributes:l}=e;return(0,g.jsx)("li",{role:"tab",tabIndex:r===n?0:-1,"aria-selected":r===n,ref:e=>{i.push(e)},onKeyDown:d,onClick:u,...l,className:(0,a.A)("tabs__item",x.tabItem,l?.className,{"tabs__item--active":r===n}),children:t??n},n)}))})}function j(e){let{lazy:n,children:t,selectedValue:l}=e;const s=(Array.isArray(t)?t:[t]).filter(Boolean);if(n){const e=s.find((e=>e.props.value===l));return e?(0,r.cloneElement)(e,{className:(0,a.A)("margin-top--md",e.props.className)}):null}return(0,g.jsx)("div",{className:"margin-top--md",children:s.map(((e,n)=>(0,r.cloneElement)(e,{key:n,hidden:e.props.value!==l})))})}function y(e){const n=p(e);return(0,g.jsxs)("div",{className:(0,a.A)("tabs-container",x.tabList),children:[(0,g.jsx)(v,{...n,...e}),(0,g.jsx)(j,{...n,...e})]})}function V(e){const n=(0,f.A)();return(0,g.jsx)(y,{...e,children:d(e.children)},String(n))}},7178:(e,n,t)=>{t.d(n,{A:()=>s});t(4700);var r=t(3372);const a={tabItem:"tabItem_Sb6x"};var l=t(7968);function s(e){let{children:n,hidden:t,className:s}=e;return(0,l.jsx)("div",{role:"tabpanel",className:(0,r.A)(a.tabItem,s),hidden:t,children:n})}}}]);