"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[324],{2892:(e,t,n)=>{n.d(t,{A:()=>l});const l=n.p+"assets/images/console-c536858e23fca5370f12f42a5341e4d0.png"},3508:(e,t,n)=>{n.d(t,{A:()=>l});const l=n.p+"assets/images/ftldev-d63e6c344de8480c77ade1b6a0b7a4f2.png"},4036:(e,t,n)=>{n.d(t,{A:()=>w});var l=n(8225),a=n(3372),r=n(8086),o=n(1654),i=n(4699),s=n(5531),c=n(2136),d=n(6070);function h(e){return l.Children.toArray(e).filter((e=>"\n"!==e)).map((e=>{if(!e||(0,l.isValidElement)(e)&&function(e){const{props:t}=e;return!!t&&"object"==typeof t&&"value"in t}(e))return e;throw new Error(`Docusaurus error: Bad <Tabs> child <${"string"==typeof e.type?e.type:e.type.name}>: all children of the <Tabs> component should be <TabItem>, and every <TabItem> should have a unique "value" prop.`)}))?.filter(Boolean)??[]}function u(e){const{values:t,children:n}=e;return(0,l.useMemo)((()=>{const e=t??function(e){return h(e).map((e=>{let{props:{value:t,label:n,attributes:l,default:a}}=e;return{value:t,label:n,attributes:l,default:a}}))}(n);return function(e){const t=(0,c.XI)(e,((e,t)=>e.value===t.value));if(t.length>0)throw new Error(`Docusaurus error: Duplicate values "${t.map((e=>e.value)).join(", ")}" found in <Tabs>. Every value needs to be unique.`)}(e),e}),[t,n])}function p(e){let{value:t,tabValues:n}=e;return n.some((e=>e.value===t))}function m(e){let{queryString:t=!1,groupId:n}=e;const a=(0,o.W6)(),r=function(e){let{queryString:t=!1,groupId:n}=e;if("string"==typeof t)return t;if(!1===t)return null;if(!0===t&&!n)throw new Error('Docusaurus error: The <Tabs> component groupId prop is required if queryString=true, because this value is used as the search param name. You can also provide an explicit value such as queryString="my-search-param".');return n??null}({queryString:t,groupId:n});return[(0,s.aZ)(r),(0,l.useCallback)((e=>{if(!r)return;const t=new URLSearchParams(a.location.search);t.set(r,e),a.replace({...a.location,search:t.toString()})}),[r,a])]}function g(e){const{defaultValue:t,queryString:n=!1,groupId:a}=e,r=u(e),[o,s]=(0,l.useState)((()=>function(e){let{defaultValue:t,tabValues:n}=e;if(0===n.length)throw new Error("Docusaurus error: the <Tabs> component requires at least one <TabItem> children component");if(t){if(!p({value:t,tabValues:n}))throw new Error(`Docusaurus error: The <Tabs> has a defaultValue "${t}" but none of its children has the corresponding value. Available values are: ${n.map((e=>e.value)).join(", ")}. If you intend to show no default tab, use defaultValue={null} instead.`);return t}const l=n.find((e=>e.default))??n[0];if(!l)throw new Error("Unexpected error: 0 tabValues");return l.value}({defaultValue:t,tabValues:r}))),[c,h]=m({queryString:n,groupId:a}),[g,f]=function(e){let{groupId:t}=e;const n=function(e){return e?`docusaurus.tab.${e}`:null}(t),[a,r]=(0,d.Dv)(n);return[a,(0,l.useCallback)((e=>{n&&r.set(e)}),[n,r])]}({groupId:a}),x=(()=>{const e=c??g;return p({value:e,tabValues:r})?e:null})();(0,i.A)((()=>{x&&s(x)}),[x]);return{selectedValue:o,selectValue:(0,l.useCallback)((e=>{if(!p({value:e,tabValues:r}))throw new Error(`Can't select invalid tab value=${e}`);s(e),h(e),f(e)}),[h,f,r]),tabValues:r}}var f=n(4185);const x={tabList:"tabList_W6YW",tabItem:"tabItem_G0iJ"};var b=n(7557);function j(e){let{className:t,block:n,selectedValue:l,selectValue:o,tabValues:i}=e;const s=[],{blockElementScrollPositionUntilNextRender:c}=(0,r.a_)(),d=e=>{const t=e.currentTarget,n=s.indexOf(t),a=i[n].value;a!==l&&(c(t),o(a))},h=e=>{let t=null;switch(e.key){case"Enter":d(e);break;case"ArrowRight":{const n=s.indexOf(e.currentTarget)+1;t=s[n]??s[0];break}case"ArrowLeft":{const n=s.indexOf(e.currentTarget)-1;t=s[n]??s[s.length-1];break}}t?.focus()};return(0,b.jsx)("ul",{role:"tablist","aria-orientation":"horizontal",className:(0,a.A)("tabs",{"tabs--block":n},t),children:i.map((e=>{let{value:t,label:n,attributes:r}=e;return(0,b.jsx)("li",{role:"tab",tabIndex:l===t?0:-1,"aria-selected":l===t,ref:e=>{s.push(e)},onKeyDown:h,onClick:d,...r,className:(0,a.A)("tabs__item",x.tabItem,r?.className,{"tabs__item--active":l===t}),children:n??t},t)}))})}function v(e){let{lazy:t,children:n,selectedValue:r}=e;const o=(Array.isArray(n)?n:[n]).filter(Boolean);if(t){const e=o.find((e=>e.props.value===r));return e?(0,l.cloneElement)(e,{className:(0,a.A)("margin-top--md",e.props.className)}):null}return(0,b.jsx)("div",{className:"margin-top--md",children:o.map(((e,t)=>(0,l.cloneElement)(e,{key:t,hidden:e.props.value!==r})))})}function y(e){const t=g(e);return(0,b.jsxs)("div",{className:(0,a.A)("tabs-container",x.tabList),children:[(0,b.jsx)(j,{...t,...e}),(0,b.jsx)(v,{...t,...e})]})}function w(e){const t=(0,f.A)();return(0,b.jsx)(y,{...e,children:h(e.children)},String(t))}},5769:(e,t,n)=>{n.d(t,{A:()=>l});const l=n.p+"assets/images/consoletrace-c201f9652fe746afcd0f902483414995.png"},6039:(e,t,n)=>{n.d(t,{R:()=>o,x:()=>i});var l=n(8225);const a={},r=l.createContext(a);function o(e){const t=l.useContext(r);return l.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function i(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:o(e.components),l.createElement(r.Provider,{value:t},e.children)}},8349:(e,t,n)=>{n.d(t,{A:()=>l});const l=n.p+"assets/images/ftlcall-e355cbec6c94c1f68c193239666acf8d.png"},8784:(e,t,n)=>{n.d(t,{A:()=>o});n(8225);var l=n(3372);const a={tabItem:"tabItem_f5BR"};var r=n(7557);function o(e){let{children:t,hidden:n,className:o}=e;return(0,r.jsx)("div",{role:"tabpanel",className:(0,l.A)(a.tabItem,o),hidden:n,children:t})}},9535:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>d,contentTitle:()=>c,default:()=>p,frontMatter:()=>s,metadata:()=>l,toc:()=>h});const l=JSON.parse('{"id":"getting-started/quick-start","title":"Quick Start","description":"One page summary of how to start a new FTL project","source":"@site/docs/getting-started/quick-start.md","sourceDirName":"getting-started","slug":"/getting-started/quick-start","permalink":"/ftl/docs/getting-started/quick-start","draft":false,"unlisted":false,"editUrl":"https://github.com/block/ftl/tree/main/docs/docs/getting-started/quick-start.md","tags":[],"version":"current","sidebarPosition":1,"frontMatter":{"sidebar_position":1,"title":"Quick Start","description":"One page summary of how to start a new FTL project"},"sidebar":"tutorialSidebar","previous":{"title":"Getting Started","permalink":"/ftl/docs/category/getting-started"},"next":{"title":"Reference","permalink":"/ftl/docs/category/reference"}}');var a=n(7557),r=n(6039),o=n(4036),i=n(8784);const s={sidebar_position:1,title:"Quick Start",description:"One page summary of how to start a new FTL project"},c="Quick Start",d={},h=[{value:"Requirements",id:"requirements",level:2},{value:"Install the FTL CLI",id:"install-the-ftl-cli",level:3},{value:"Install the VSCode extension",id:"install-the-vscode-extension",level:3},{value:"Development",id:"development",level:2},{value:"Initialize an FTL project",id:"initialize-an-ftl-project",level:3},{value:"Create a new module",id:"create-a-new-module",level:3},{value:"Start the FTL cluster",id:"start-the-ftl-cluster",level:3},{value:"Open the console",id:"open-the-console",level:3},{value:"Call your verb",id:"call-your-verb",level:3},{value:"Create another module",id:"create-another-module",level:3}];function u(e){const t={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",header:"header",img:"img",p:"p",pre:"pre",...(0,r.R)(),...e.components};return(0,a.jsxs)(a.Fragment,{children:[(0,a.jsx)(t.header,{children:(0,a.jsx)(t.h1,{id:"quick-start",children:"Quick Start"})}),"\n",(0,a.jsx)(t.p,{children:"One page summary of how to start a new FTL project."}),"\n",(0,a.jsx)(t.h2,{id:"requirements",children:"Requirements"}),"\n",(0,a.jsx)(t.h3,{id:"install-the-ftl-cli",children:"Install the FTL CLI"}),"\n",(0,a.jsxs)(t.p,{children:["Install the FTL CLI on Mac or Linux via ",(0,a.jsx)(t.a,{href:"https://brew.sh/",children:"Homebrew"}),", ",(0,a.jsx)(t.a,{href:"https://cashapp.github.io/hermit",children:"Hermit"}),", or manually."]}),"\n","\n",(0,a.jsxs)(o.A,{groupId:"package-manager",children:[(0,a.jsx)(i.A,{value:"homebrew",label:"Homebrew",default:!0,children:(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-bash",children:"brew tap block/ftl && brew install ftl\n"})})}),(0,a.jsxs)(i.A,{value:"hermit",label:"Hermit",children:[(0,a.jsx)(t.p,{children:"FTL can be installed from the main Hermit package repository by simply:"}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-bash",children:"hermit install ftl\n"})}),(0,a.jsxs)(t.p,{children:["Alternatively you can add ",(0,a.jsx)(t.a,{href:"https://github.com/block/hermit-ftl",children:"hermit-ftl"})," to your sources by adding the following to your Hermit environment's ",(0,a.jsx)(t.code,{children:"bin/hermit.hcl"})," file:"]}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-hcl",children:'sources = ["https://github.com/block/hermit-ftl.git", "https://github.com/cashapp/hermit-packages.git"]\n'})})]}),(0,a.jsx)(i.A,{value:"manual",label:"Manually",children:(0,a.jsxs)(t.p,{children:["Download binaries from the ",(0,a.jsx)(t.a,{href:"https://github.com/block/ftl/releases/latest",children:"latest release page"})," and place them in your ",(0,a.jsx)(t.code,{children:"$PATH"}),"."]})})]}),"\n",(0,a.jsx)(t.h3,{id:"install-the-vscode-extension",children:"Install the VSCode extension"}),"\n",(0,a.jsxs)(t.p,{children:["The ",(0,a.jsx)(t.a,{href:"https://marketplace.visualstudio.com/items?itemName=FTL.ftl",children:"FTL VSCode extension"})," provides error and problem reporting through the language server and includes code snippets for common FTL patterns."]}),"\n",(0,a.jsx)(t.h2,{id:"development",children:"Development"}),"\n",(0,a.jsx)(t.h3,{id:"initialize-an-ftl-project",children:"Initialize an FTL project"}),"\n",(0,a.jsx)(t.p,{children:"Once FTL is installed, initialize an FTL project:"}),"\n",(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-bash",children:"ftl init myproject\ncd myproject\n"})}),"\n",(0,a.jsxs)(t.p,{children:["This will create a new ",(0,a.jsx)(t.code,{children:"myproject"})," directory containing an ",(0,a.jsx)(t.code,{children:"ftl-project.toml"})," file, a git repository, and a ",(0,a.jsx)(t.code,{children:"bin/"})," directory with Hermit tooling. The Hermit tooling includes the current version of FTL, and language support for go and JVM based languages."]}),"\n",(0,a.jsx)(t.h3,{id:"create-a-new-module",children:"Create a new module"}),"\n",(0,a.jsx)(t.p,{children:"Now that you have an FTL project, create a new module:"}),"\n",(0,a.jsxs)(o.A,{groupId:"languages",children:[(0,a.jsxs)(i.A,{value:"go",label:"Go",default:!0,children:[(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-bash",children:"ftl module new go alice\n"})}),(0,a.jsxs)(t.p,{children:["This will place the code for the new module ",(0,a.jsx)(t.code,{children:"alice"})," in ",(0,a.jsx)(t.code,{children:"myproject/alice/alice.go"}),":"]}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-go",children:'package alice\n\nimport (\n    "context"\n    "fmt"\n\n    "github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.\n)\n\n//ftl:verb\nfunc Echo(ctx context.Context, name ftl.Option[string]) (string, error) {\n    return fmt.Sprintf("Hello, %s!", name.Default("anonymous")), nil\n}\n'})}),(0,a.jsx)(t.p,{children:"Each module is its own Go module."})]}),(0,a.jsxs)(i.A,{value:"kotlin",label:"Kotlin",children:[(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-bash",children:"ftl module new kotlin alice\n"})}),(0,a.jsxs)(t.p,{children:["This will create a new Maven ",(0,a.jsx)(t.code,{children:"pom.xml"})," based project in the directory ",(0,a.jsx)(t.code,{children:"alice"})," and create new example code in ",(0,a.jsx)(t.code,{children:"alice/src/main/kotlin/ftl/alice/Alice.kt"}),":"]}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-kotlin",children:'package com.example\n\nimport xyz.block.ftl.Export\nimport xyz.block.ftl.Verb\n\n\n@Export\n@Verb\nfun hello(req: String): String = "Hello, $req!"\n'})})]}),(0,a.jsxs)(i.A,{value:"java",label:"Java",children:[(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-bash",children:"ftl module new java alice\n"})}),(0,a.jsxs)(t.p,{children:["This will create a new Maven ",(0,a.jsx)(t.code,{children:"pom.xml"})," based project in the directory ",(0,a.jsx)(t.code,{children:"alice"})," and create new example code in ",(0,a.jsx)(t.code,{children:"alice/src/main/java/ftl/alice/Alice.java"}),":"]}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-java",children:'package com.example;\n\nimport xyz.block.ftl.Export;\nimport xyz.block.ftl.Verb;\n\npublic class Alice {\n\n    @Export\n    @Verb\n    public String hello(String request) {\n        return "Hello, " + request + "!";\n    }\n}\n'})})]}),(0,a.jsxs)(i.A,{value:"schema",label:"Schema",children:[(0,a.jsx)(t.p,{children:"When you create a new module, FTL generates a schema that represents your code. For the examples above, the schema would look like:"}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-schema",children:"module alice {\n  verb echo(String?) String\n}\n"})}),(0,a.jsx)(t.p,{children:"The schema is automatically generated from your code and represents the structure of your FTL module, including data types, verbs, and their relationships."})]})]}),"\n",(0,a.jsx)(t.p,{children:"Any number of modules can be added to your project, adjacent to each other."}),"\n",(0,a.jsx)(t.h3,{id:"start-the-ftl-cluster",children:"Start the FTL cluster"}),"\n",(0,a.jsx)(t.p,{children:"Start the local FTL development cluster from the command-line:"}),"\n",(0,a.jsx)(t.p,{children:(0,a.jsx)(t.img,{alt:"ftl dev",src:n(3508).A+"",width:"1364",height:"990"})}),"\n",(0,a.jsxs)(t.p,{children:["This will build and deploy all local modules. Modifying the code will cause ",(0,a.jsx)(t.code,{children:"ftl dev"})," to rebuild and redeploy the module."]}),"\n",(0,a.jsx)(t.h3,{id:"open-the-console",children:"Open the console"}),"\n",(0,a.jsxs)(t.p,{children:["FTL has a console that allows interaction with the cluster topology, logs, traces, and more. Open a browser window at ",(0,a.jsx)(t.a,{href:"http://localhost:8899",children:"http://localhost:8899"})," to view it:"]}),"\n",(0,a.jsx)(t.p,{children:(0,a.jsx)(t.img,{alt:"FTL Console",src:n(2892).A+"",width:"2344",height:"1802"})}),"\n",(0,a.jsx)(t.h3,{id:"call-your-verb",children:"Call your verb"}),"\n",(0,a.jsx)(t.p,{children:"You can call verbs from the console:"}),"\n",(0,a.jsx)(t.p,{children:(0,a.jsx)(t.img,{alt:"console call",src:n(9664).A+"",width:"2432",height:"1890"})}),"\n",(0,a.jsxs)(t.p,{children:["Or from a terminal use ",(0,a.jsx)(t.code,{children:"ftl call"})," to call your verb:"]}),"\n",(0,a.jsx)(t.p,{children:(0,a.jsx)(t.img,{alt:"ftl call",src:n(8349).A+"",width:"1364",height:"990"})}),"\n",(0,a.jsx)(t.p,{children:"And view your trace in the console:"}),"\n",(0,a.jsx)(t.p,{children:(0,a.jsx)(t.img,{alt:"console trace",src:n(5769).A+"",width:"2432",height:"1890"})}),"\n",(0,a.jsx)(t.h3,{id:"create-another-module",children:"Create another module"}),"\n",(0,a.jsxs)(t.p,{children:["Create another module and call ",(0,a.jsx)(t.code,{children:"alice.echo"})," from it with by importing the ",(0,a.jsx)(t.code,{children:"alice"})," module and adding the verb client, ",(0,a.jsx)(t.code,{children:"alice.EchoClient"}),", to the signature of the calling verb. It can be invoked as a function:"]}),"\n",(0,a.jsxs)(o.A,{groupId:"languages",children:[(0,a.jsx)(i.A,{value:"go",label:"Go",default:!0,children:(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-go",children:'//ftl:verb\nimport "ftl/alice"\n\n//ftl:verb\nfunc Other(ctx context.Context, in string, ec alice.EchoClient) (string, error) {\n    out, err := ec(ctx, in)\n    ...\n}\n'})})}),(0,a.jsxs)(i.A,{value:"kotlin",label:"Kotlin",children:[(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-kotlin",children:'package com.example\n\nimport xyz.block.ftl.Export\nimport xyz.block.ftl.Verb\nimport ftl.alice.EchoClient\n\n\n@Export\n@Verb\nfun other(req: String, echo: EchoClient): String = "Hello from Other , ${echo.call(req)}!"\n'})}),(0,a.jsxs)(t.p,{children:["Note that the ",(0,a.jsx)(t.code,{children:"EchoClient"})," is generated by FTL and must be imported. Unfortunately at the moment JVM based languages have a bit of a chicken-and-egg problem with the generated clients. To force a dependency between the modules you need to add an import on a class that does not exist yet, and then FTL will generate the client for you. This will be fixed in the future."]})]}),(0,a.jsxs)(i.A,{value:"java",label:"Java",children:[(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-java",children:'package com.example.client;\n\nimport xyz.block.ftl.Export;\nimport xyz.block.ftl.Verb;\nimport ftl.alice.EchoClient;\n\npublic class OtherVerb {\n\n    @Export\n    @Verb\n    public String other(String request, EchoClient echoClient) {\n        return "Hello, " + echoClient.call(request) + "!";\n    }\n}\n'})}),(0,a.jsxs)(t.p,{children:["Note that the ",(0,a.jsx)(t.code,{children:"EchoClient"})," is generated by FTL and must be imported. Unfortunately at the moment JVM based languages have a bit of a chicken-and-egg problem with the generated clients. To force a dependency between the modules you need to add an import on a class that does not exist yet, and then FTL will generate the client for you. This will be fixed in the future."]})]}),(0,a.jsxs)(i.A,{value:"schema",label:"Schema",children:[(0,a.jsx)(t.p,{children:"When you create a second module that calls the first one, the schema would look like:"}),(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-schema",children:"module alice {\n  export verb echo(String?) String\n}\n\nmodule other {\n  export verb other(String) String\n    +calls alice.echo\n}\n"})}),(0,a.jsxs)(t.p,{children:["The ",(0,a.jsx)(t.code,{children:"+calls"})," annotation in the schema indicates that the ",(0,a.jsx)(t.code,{children:"other"})," verb calls the ",(0,a.jsx)(t.code,{children:"echo"})," verb from the ",(0,a.jsx)(t.code,{children:"alice"})," module."]})]})]})]})}function p(e={}){const{wrapper:t}={...(0,r.R)(),...e.components};return t?(0,a.jsx)(t,{...e,children:(0,a.jsx)(u,{...e})}):u(e)}},9664:(e,t,n)=>{n.d(t,{A:()=>l});const l=n.p+"assets/images/consolecall-c9a49fb473e7fdf16a210ed7d4257ec4.png"}}]);