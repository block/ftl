"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[828],{7417:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>u,contentTitle:()=>c,default:()=>b,frontMatter:()=>i,metadata:()=>r,toc:()=>d});const r=JSON.parse('{"id":"reference/verbs","title":"Verbs","description":"Declaring and calling Verbs","source":"@site/docs/reference/verbs.md","sourceDirName":"reference","slug":"/reference/verbs","permalink":"/ftl/docs/reference/verbs","draft":false,"unlisted":false,"editUrl":"https://github.com/block/ftl/tree/main/docs/docs/reference/verbs.md","tags":[],"version":"current","sidebarPosition":1,"frontMatter":{"sidebar_position":1,"title":"Verbs","description":"Declaring and calling Verbs"},"sidebar":"tutorialSidebar","previous":{"title":"Reference","permalink":"/ftl/docs/category/reference"},"next":{"title":"Types","permalink":"/ftl/docs/reference/types"}}');var l=t(7557),a=t(972),s=t(8630),o=t(4932);const i={sidebar_position:1,title:"Verbs",description:"Declaring and calling Verbs"},c="Verbs",u={},d=[{value:"Defining Verbs",id:"defining-verbs",level:2},{value:"Calling Verbs",id:"calling-verbs",level:2}];function h(e){const n={a:"a",code:"code",h1:"h1",h2:"h2",header:"header",p:"p",pre:"pre",...(0,a.R)(),...e.components};return(0,l.jsxs)(l.Fragment,{children:[(0,l.jsx)(n.header,{children:(0,l.jsx)(n.h1,{id:"verbs",children:"Verbs"})}),"\n",(0,l.jsx)(n.h2,{id:"defining-verbs",children:"Defining Verbs"}),"\n","\n",(0,l.jsxs)(s.A,{groupId:"languages",children:[(0,l.jsxs)(o.A,{value:"go",label:"Go",default:!0,children:[(0,l.jsxs)(n.p,{children:["To declare a Verb, write a normal Go function with the following signature, annotated with the Go ",(0,l.jsx)(n.a,{href:"https://tip.golang.org/doc/comment#syntax",children:"comment directive"})," ",(0,l.jsx)(n.code,{children:"//ftl:verb"}),":"]}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-go",children:"//ftl:verb\nfunc F(context.Context, In) (Out, error) { }\n"})}),(0,l.jsx)(n.p,{children:"eg."}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-go",children:"type EchoRequest struct {}\n\ntype EchoResponse struct {}\n\n//ftl:verb\nfunc Echo(ctx context.Context, in EchoRequest) (EchoResponse, error) {\n  // ...\n}\n"})})]}),(0,l.jsxs)(o.A,{value:"kotlin",label:"Kotlin",children:[(0,l.jsxs)(n.p,{children:["To declare a Verb, write a normal Kotlin function with the following signature, annotated with the Kotlin ",(0,l.jsx)(n.a,{href:"https://kotlinlang.org/docs/annotations.html",children:"annotation"})," ",(0,l.jsx)(n.code,{children:"@Verb"}),":"]}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-kotlin",children:"@Verb\nfun F(In): Out { }\n"})}),(0,l.jsx)(n.p,{children:"eg."}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-kotlin",children:"data class EchoRequest\ndata class EchoResponse\n\n@Verb\nfun echo(request: EchoRequest): EchoResponse {\n  // ...\n}\n"})})]}),(0,l.jsxs)(o.A,{value:"java",label:"Java",children:[(0,l.jsxs)(n.p,{children:["To declare a Verb, write a normal Java method with the following signature, annotated with the ",(0,l.jsx)(n.code,{children:"@Verb"})," annotation:"]}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-java",children:"@Verb\npublic Output f(Input input) { }\n"})}),(0,l.jsx)(n.p,{children:"eg."}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-java",children:"import xyz.block.ftl.Verb;\n\nclass EchoRequest {}\n\nclass EchoResponse {}\n\npublic class EchoClass {\n    @Verb\n    public EchoResponse echo(EchoRequest request) {\n        // ...\n    }\n}\n"})})]})]}),"\n",(0,l.jsxs)(n.p,{children:["By default verbs are only visible to other verbs in the same module (see ",(0,l.jsx)(n.a,{href:"./visibility",children:"visibility"})," for more information)."]}),"\n",(0,l.jsx)(n.h2,{id:"calling-verbs",children:"Calling Verbs"}),"\n",(0,l.jsxs)(s.A,{groupId:"languages",children:[(0,l.jsxs)(o.A,{value:"go",label:"Go",default:!0,children:[(0,l.jsxs)(n.p,{children:["To call a verb, import the module's verb client (",(0,l.jsx)(n.code,{children:"{ModuleName}.{VerbName}Client"}),"), add it to your verb's signature, then invoke it as a function. eg."]}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-go",children:"//ftl:verb\nfunc Echo(ctx context.Context, in EchoRequest, tc time.TimeClient) (EchoResponse, error) {\n    out, err := tc(ctx, TimeRequest{...})\n}\n"})}),(0,l.jsxs)(n.p,{children:["Verb clients are generated by FTL. If the callee verb belongs to the same module as the caller, you must build the\nmodule first (with callee verb defined) in order to generate its client for use by the caller. Local verb clients are\navailable in the generated ",(0,l.jsx)(n.code,{children:"types.ftl.go"})," file as ",(0,l.jsx)(n.code,{children:"{VerbName}Client"}),"."]})]}),(0,l.jsxs)(o.A,{value:"kotlin",label:"Kotlin",children:[(0,l.jsxs)(n.p,{children:["To call a verb, import the module's verb client, add it to your verb's signature, then ",(0,l.jsx)(n.code,{children:"call()"})," it. eg."]}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-kotlin",children:"import ftl.time.TimeClient\nimport xyz.block.ftl.Verb\n\n@Verb\nfun echo(req: EchoRequest, time: TimeClient): EchoResponse {\n  val response = time.call()\n  // ...\n}\n\nval response = time.call()\n"})}),(0,l.jsx)(n.p,{children:"Verb clients are generated by FTL. If the callee verb belongs to the same module as the caller, you must manually define your\nown client:"}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-kotlin",children:'@VerbClient(name="time")\ninterface TimeClient {\n    fun call(): TimeResponse\n}\n'})})]}),(0,l.jsxs)(o.A,{value:"java",label:"Java",children:[(0,l.jsx)(n.p,{children:"To call a verb, import the module's verb client, add it to your verb's signature, then call it. eg."}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-java",children:"import ftl.time.TimeClient;\nimport xyz.block.ftl.Verb;\n\npublic class EchoClass {\n    @Verb\n    public EchoResponse echo(EchoRequest request, TimeClient time) {\n        TimeResponse response = time.call();\n        // ...\n    }\n}\n"})}),(0,l.jsx)(n.p,{children:"Verb clients are generated by FTL. If the callee verb belongs to the same module as the caller, you must manually define your\nown client:"}),(0,l.jsx)(n.pre,{children:(0,l.jsx)(n.code,{className:"language-java",children:'@VerbClient(name="time")\npublic interface TimeClient {\n    TimeResponse call();\n}\n'})})]})]})]})}function b(e={}){const{wrapper:n}={...(0,a.R)(),...e.components};return n?(0,l.jsx)(n,{...e,children:(0,l.jsx)(h,{...e})}):h(e)}},4932:(e,n,t)=>{t.d(n,{A:()=>s});t(8225);var r=t(3372);const l={tabItem:"tabItem_tr6E"};var a=t(7557);function s(e){let{children:n,hidden:t,className:s}=e;return(0,a.jsx)("div",{role:"tabpanel",className:(0,r.A)(l.tabItem,s),hidden:t,children:n})}},8630:(e,n,t)=>{t.d(n,{A:()=>V});var r=t(8225),l=t(3372),a=t(1910),s=t(1654),o=t(932),i=t(2955),c=t(1192),u=t(7931);function d(e){return r.Children.toArray(e).filter((e=>"\n"!==e)).map((e=>{if(!e||(0,r.isValidElement)(e)&&function(e){const{props:n}=e;return!!n&&"object"==typeof n&&"value"in n}(e))return e;throw new Error(`Docusaurus error: Bad <Tabs> child <${"string"==typeof e.type?e.type:e.type.name}>: all children of the <Tabs> component should be <TabItem>, and every <TabItem> should have a unique "value" prop.`)}))?.filter(Boolean)??[]}function h(e){const{values:n,children:t}=e;return(0,r.useMemo)((()=>{const e=n??function(e){return d(e).map((e=>{let{props:{value:n,label:t,attributes:r,default:l}}=e;return{value:n,label:t,attributes:r,default:l}}))}(t);return function(e){const n=(0,c.XI)(e,((e,n)=>e.value===n.value));if(n.length>0)throw new Error(`Docusaurus error: Duplicate values "${n.map((e=>e.value)).join(", ")}" found in <Tabs>. Every value needs to be unique.`)}(e),e}),[n,t])}function b(e){let{value:n,tabValues:t}=e;return t.some((e=>e.value===n))}function p(e){let{queryString:n=!1,groupId:t}=e;const l=(0,s.W6)(),a=function(e){let{queryString:n=!1,groupId:t}=e;if("string"==typeof n)return n;if(!1===n)return null;if(!0===n&&!t)throw new Error('Docusaurus error: The <Tabs> component groupId prop is required if queryString=true, because this value is used as the search param name. You can also provide an explicit value such as queryString="my-search-param".');return t??null}({queryString:n,groupId:t});return[(0,i.aZ)(a),(0,r.useCallback)((e=>{if(!a)return;const n=new URLSearchParams(l.location.search);n.set(a,e),l.replace({...l.location,search:n.toString()})}),[a,l])]}function m(e){const{defaultValue:n,queryString:t=!1,groupId:l}=e,a=h(e),[s,i]=(0,r.useState)((()=>function(e){let{defaultValue:n,tabValues:t}=e;if(0===t.length)throw new Error("Docusaurus error: the <Tabs> component requires at least one <TabItem> children component");if(n){if(!b({value:n,tabValues:t}))throw new Error(`Docusaurus error: The <Tabs> has a defaultValue "${n}" but none of its children has the corresponding value. Available values are: ${t.map((e=>e.value)).join(", ")}. If you intend to show no default tab, use defaultValue={null} instead.`);return n}const r=t.find((e=>e.default))??t[0];if(!r)throw new Error("Unexpected error: 0 tabValues");return r.value}({defaultValue:n,tabValues:a}))),[c,d]=p({queryString:t,groupId:l}),[m,f]=function(e){let{groupId:n}=e;const t=function(e){return e?`docusaurus.tab.${e}`:null}(n),[l,a]=(0,u.Dv)(t);return[l,(0,r.useCallback)((e=>{t&&a.set(e)}),[t,a])]}({groupId:l}),g=(()=>{const e=c??m;return b({value:e,tabValues:a})?e:null})();(0,o.A)((()=>{g&&i(g)}),[g]);return{selectedValue:s,selectValue:(0,r.useCallback)((e=>{if(!b({value:e,tabValues:a}))throw new Error(`Can't select invalid tab value=${e}`);i(e),d(e),f(e)}),[d,f,a]),tabValues:a}}var f=t(224);const g={tabList:"tabList_t6iw",tabItem:"tabItem_FCC1"};var v=t(7557);function x(e){let{className:n,block:t,selectedValue:r,selectValue:s,tabValues:o}=e;const i=[],{blockElementScrollPositionUntilNextRender:c}=(0,a.a_)(),u=e=>{const n=e.currentTarget,t=i.indexOf(n),l=o[t].value;l!==r&&(c(n),s(l))},d=e=>{let n=null;switch(e.key){case"Enter":u(e);break;case"ArrowRight":{const t=i.indexOf(e.currentTarget)+1;n=i[t]??i[0];break}case"ArrowLeft":{const t=i.indexOf(e.currentTarget)-1;n=i[t]??i[i.length-1];break}}n?.focus()};return(0,v.jsx)("ul",{role:"tablist","aria-orientation":"horizontal",className:(0,l.A)("tabs",{"tabs--block":t},n),children:o.map((e=>{let{value:n,label:t,attributes:a}=e;return(0,v.jsx)("li",{role:"tab",tabIndex:r===n?0:-1,"aria-selected":r===n,ref:e=>{i.push(e)},onKeyDown:d,onClick:u,...a,className:(0,l.A)("tabs__item",g.tabItem,a?.className,{"tabs__item--active":r===n}),children:t??n},n)}))})}function j(e){let{lazy:n,children:t,selectedValue:a}=e;const s=(Array.isArray(t)?t:[t]).filter(Boolean);if(n){const e=s.find((e=>e.props.value===a));return e?(0,r.cloneElement)(e,{className:(0,l.A)("margin-top--md",e.props.className)}):null}return(0,v.jsx)("div",{className:"margin-top--md",children:s.map(((e,n)=>(0,r.cloneElement)(e,{key:n,hidden:e.props.value!==a})))})}function y(e){const n=m(e);return(0,v.jsxs)("div",{className:(0,l.A)("tabs-container",g.tabList),children:[(0,v.jsx)(x,{...n,...e}),(0,v.jsx)(j,{...n,...e})]})}function V(e){const n=(0,f.A)();return(0,v.jsx)(y,{...e,children:d(e.children)},String(n))}},972:(e,n,t)=>{t.d(n,{R:()=>s,x:()=>o});var r=t(8225);const l={},a=r.createContext(l);function s(e){const n=r.useContext(a);return r.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function o(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(l):e.components||l:s(e.components),r.createElement(a.Provider,{value:n},e.children)}}}]);