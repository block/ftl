"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[278],{1407:(e,n,t)=>{t.d(n,{A:()=>l});t(8225);var a=t(3372);const r={tabItem:"tabItem_LXtO"};var s=t(7557);function l(e){let{children:n,hidden:t,className:l}=e;return(0,s.jsx)("div",{role:"tabpanel",className:(0,a.A)(r.tabItem,l),hidden:t,children:n})}},3455:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>u,contentTitle:()=>c,default:()=>h,frontMatter:()=>o,metadata:()=>a,toc:()=>p});const a=JSON.parse('{"id":"reference/externaltypes","title":"External Types","description":"Using external types in your modules","source":"@site/docs/reference/externaltypes.md","sourceDirName":"reference","slug":"/reference/externaltypes","permalink":"/ftl/docs/reference/externaltypes","draft":false,"unlisted":false,"editUrl":"https://github.com/block/ftl/tree/main/docs/docs/reference/externaltypes.md","tags":[],"version":"current","sidebarPosition":16,"frontMatter":{"sidebar_position":16,"title":"External Types","description":"Using external types in your modules"},"sidebar":"tutorialSidebar","previous":{"title":"HTTP Ingress","permalink":"/ftl/docs/reference/ingress"},"next":{"title":"Databases","permalink":"/ftl/docs/reference/databases"}}');var r=t(7557),s=t(7389),l=t(9077),i=t(1407);const o={sidebar_position:16,title:"External Types",description:"Using external types in your modules"},c="External Types",u={},p=[];function d(e){const n={code:"code",h1:"h1",header:"header",li:"li",ol:"ol",p:"p",pre:"pre",ul:"ul",...(0,s.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(n.header,{children:(0,r.jsx)(n.h1,{id:"external-types",children:"External Types"})}),"\n",(0,r.jsx)(n.p,{children:"FTL supports the use of external types in your FTL modules. External types are types defined in other packages or modules that are not part of the FTL module."}),"\n",(0,r.jsx)(n.p,{children:"The primary difference is that external types are not defined in the FTL schema, and therefore serialization and deserialization of these types is not handled\nby FTL. Instead, FTL relies on the runtime to handle serialization and deserialization of these types."}),"\n",(0,r.jsx)(n.p,{children:"In some cases this feature can also be used to provide custom serialization and deserialization logic for types that are not directly supported by FTL, even\nif they are defined in the same package as the FTL module."}),"\n",(0,r.jsx)(n.p,{children:"When using external types:"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:["The external type is typically widened to ",(0,r.jsx)(n.code,{children:"Any"})," in the FTL schema by default"]}),"\n",(0,r.jsxs)(n.li,{children:["You can map to specific FTL types (like ",(0,r.jsx)(n.code,{children:"String"}),") for better schema clarity"]}),"\n",(0,r.jsxs)(n.li,{children:["For JVM languages, ",(0,r.jsx)(n.code,{children:"java"})," is always used as the runtime name in the schema, regardless of whether Kotlin or Java is used"]}),"\n",(0,r.jsxs)(n.li,{children:["Multiple ",(0,r.jsx)(n.code,{children:"+typemap"})," annotations can be used to support cross-runtime interoperability"]}),"\n"]}),"\n","\n",(0,r.jsxs)(l.A,{groupId:"languages",children:[(0,r.jsxs)(i.A,{value:"go",label:"Go",default:!0,children:[(0,r.jsx)(n.p,{children:"To use an external type in your FTL module schema, declare a type alias over the external type:"}),(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-go",children:"//ftl:typealias\ntype FtlType external.OtherType\n\n//ftl:typealias\ntype FtlType2 = external.OtherType\n"})}),(0,r.jsx)(n.p,{children:"You can also specify mappings for other runtimes:"}),(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-go",children:'//ftl:typealias\n//ftl:typemap java "com.external.other.OtherType"\ntype FtlType external.OtherType\n'})})]}),(0,r.jsxs)(i.A,{value:"kotlin",label:"Kotlin",children:[(0,r.jsxs)(n.p,{children:["To use an external type in your FTL module schema, implement a ",(0,r.jsx)(n.code,{children:"TypeAliasMapper"}),":"]}),(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-kotlin",children:'@TypeAlias(name = "OtherType")\nclass OtherTypeTypeMapper : TypeAliasMapper<OtherType, JsonNode> {\n    override fun encode(`object`: OtherType): JsonNode {\n        return TextNode.valueOf(`object`.value)\n    }\n\n    override fun decode(serialized: JsonNode): OtherType {\n        if (serialized.isTextual) {\n            return OtherType(serialized.textValue())\n        }\n        throw RuntimeException("Expected a textual value")\n    }\n}\n'})}),(0,r.jsxs)(n.p,{children:["Note that for JVM languages ",(0,r.jsx)(n.code,{children:"java"})," is always used as the runtime name, regardless of the actual language used."]}),(0,r.jsxs)(n.p,{children:["It is also possible to map to any other valid FTL type (e.g. ",(0,r.jsx)(n.code,{children:"String"}),") by using this as the second type parameter:"]}),(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-kotlin",children:'@TypeAlias(name = "OtherType")\nclass OtherTypeTypeMapper : TypeAliasMapper<OtherType, String> {\n    override fun encode(other: OtherType): JsonNode {\n        return other.value\n    }\n\n    override fun decode(serialized: String): OtherType {\n        return OtherType(serialized.textValue())\n    }\n}\n'})}),(0,r.jsx)(n.p,{children:"You can also specify mappings for other runtimes:"}),(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-kotlin",children:'@TypeAlias(\n  name = "OtherType",\n  languageTypeMappings = [LanguageTypeMapping(language = "go", type = "github.com/external.OtherType")]\n)\n'})})]}),(0,r.jsxs)(i.A,{value:"java",label:"Java",children:[(0,r.jsxs)(n.p,{children:["To use an external type in your FTL module schema, implement a ",(0,r.jsx)(n.code,{children:"TypeAliasMapper"}),":"]}),(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-java",children:'@TypeAlias(name = "OtherType")\npublic class OtherTypeTypeMapper implements TypeAliasMapper<OtherType, JsonNode> {\n    @Override\n    public JsonNode encode(OtherType object) {\n        return TextNode.valueOf(object.getValue());\n    }\n\n    @Override\n    public AnySerializedType decode(OtherType serialized) {\n        if (serialized.isTextual()) {\n            return new OtherType(serialized.textValue());\n        }\n        throw new RuntimeException("Expected a textual value");\n    }\n}\n'})}),(0,r.jsxs)(n.p,{children:["It is also possible to map to any other valid FTL type (e.g. ",(0,r.jsx)(n.code,{children:"String"}),") by using this as the second type parameter:"]}),(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-java",children:'@TypeAlias(name = "OtherType")\npublic class OtherTypeTypeMapper implements TypeAliasMapper<OtherType, String> {\n    @Override\n    public String encode(OtherType object) {\n        return object.getValue();\n    }\n\n    @Override\n    public String decode(OtherType serialized) {\n        return new OtherType(serialized.textValue());\n    }\n}\n'})}),(0,r.jsx)(n.p,{children:"You can also specify mappings for other runtimes:"}),(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-java",children:'@TypeAlias(name = "OtherType", languageTypeMappings = {\n    @LanguageTypeMapping(language = "go", type = "github.com/external.OtherType"),\n})\n'})})]}),(0,r.jsxs)(i.A,{value:"schema",label:"Schema",children:[(0,r.jsxs)(n.p,{children:["In the FTL schema, external types are represented as type aliases with the ",(0,r.jsx)(n.code,{children:"+typemap"})," annotation:"]}),(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-schema",children:'module example {\n  // External type widened to Any\n  typealias FtlType Any\n    +typemap go "github.com/external.OtherType"\n    +typemap java "foo.bar.OtherType"\n  \n  // External type mapped to a specific FTL type\n  typealias UserID String\n    +typemap go "github.com/myapp.UserID"\n    +typemap java "com.myapp.UserID"\n  \n  // Using external types in data structures\n  data User {\n    id example.UserID\n    preferences example.FtlType\n  }\n  \n  // Using external types in verbs\n  verb processUser(example.User) Unit\n}\n'})}),(0,r.jsxs)(n.p,{children:["The ",(0,r.jsx)(n.code,{children:"+typemap"})," annotation specifies:"]}),(0,r.jsxs)(n.ol,{children:["\n",(0,r.jsx)(n.li,{children:"The target runtime/language (go, java, etc.)"}),"\n",(0,r.jsx)(n.li,{children:"The fully qualified type name in that runtime"}),"\n"]}),(0,r.jsx)(n.p,{children:"This allows FTL to decode the type properly in different languages, enabling seamless interoperability across different runtimes."})]})]})]})}function h(e={}){const{wrapper:n}={...(0,s.R)(),...e.components};return n?(0,r.jsx)(n,{...e,children:(0,r.jsx)(d,{...e})}):d(e)}},7389:(e,n,t)=>{t.d(n,{R:()=>l,x:()=>i});var a=t(8225);const r={},s=a.createContext(r);function l(e){const n=a.useContext(s);return a.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function i(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(r):e.components||r:l(e.components),a.createElement(s.Provider,{value:n},e.children)}},9077:(e,n,t)=>{t.d(n,{A:()=>v});var a=t(8225),r=t(3372),s=t(9101),l=t(1654),i=t(8827),o=t(6550),c=t(4727),u=t(6716);function p(e){return a.Children.toArray(e).filter((e=>"\n"!==e)).map((e=>{if(!e||(0,a.isValidElement)(e)&&function(e){const{props:n}=e;return!!n&&"object"==typeof n&&"value"in n}(e))return e;throw new Error(`Docusaurus error: Bad <Tabs> child <${"string"==typeof e.type?e.type:e.type.name}>: all children of the <Tabs> component should be <TabItem>, and every <TabItem> should have a unique "value" prop.`)}))?.filter(Boolean)??[]}function d(e){const{values:n,children:t}=e;return(0,a.useMemo)((()=>{const e=n??function(e){return p(e).map((e=>{let{props:{value:n,label:t,attributes:a,default:r}}=e;return{value:n,label:t,attributes:a,default:r}}))}(t);return function(e){const n=(0,c.XI)(e,((e,n)=>e.value===n.value));if(n.length>0)throw new Error(`Docusaurus error: Duplicate values "${n.map((e=>e.value)).join(", ")}" found in <Tabs>. Every value needs to be unique.`)}(e),e}),[n,t])}function h(e){let{value:n,tabValues:t}=e;return t.some((e=>e.value===n))}function y(e){let{queryString:n=!1,groupId:t}=e;const r=(0,l.W6)(),s=function(e){let{queryString:n=!1,groupId:t}=e;if("string"==typeof n)return n;if(!1===n)return null;if(!0===n&&!t)throw new Error('Docusaurus error: The <Tabs> component groupId prop is required if queryString=true, because this value is used as the search param name. You can also provide an explicit value such as queryString="my-search-param".');return t??null}({queryString:n,groupId:t});return[(0,o.aZ)(s),(0,a.useCallback)((e=>{if(!s)return;const n=new URLSearchParams(r.location.search);n.set(s,e),r.replace({...r.location,search:n.toString()})}),[s,r])]}function m(e){const{defaultValue:n,queryString:t=!1,groupId:r}=e,s=d(e),[l,o]=(0,a.useState)((()=>function(e){let{defaultValue:n,tabValues:t}=e;if(0===t.length)throw new Error("Docusaurus error: the <Tabs> component requires at least one <TabItem> children component");if(n){if(!h({value:n,tabValues:t}))throw new Error(`Docusaurus error: The <Tabs> has a defaultValue "${n}" but none of its children has the corresponding value. Available values are: ${t.map((e=>e.value)).join(", ")}. If you intend to show no default tab, use defaultValue={null} instead.`);return n}const a=t.find((e=>e.default))??t[0];if(!a)throw new Error("Unexpected error: 0 tabValues");return a.value}({defaultValue:n,tabValues:s}))),[c,p]=y({queryString:t,groupId:r}),[m,x]=function(e){let{groupId:n}=e;const t=function(e){return e?`docusaurus.tab.${e}`:null}(n),[r,s]=(0,u.Dv)(t);return[r,(0,a.useCallback)((e=>{t&&s.set(e)}),[t,s])]}({groupId:r}),f=(()=>{const e=c??m;return h({value:e,tabValues:s})?e:null})();(0,i.A)((()=>{f&&o(f)}),[f]);return{selectedValue:l,selectValue:(0,a.useCallback)((e=>{if(!h({value:e,tabValues:s}))throw new Error(`Can't select invalid tab value=${e}`);o(e),p(e),x(e)}),[p,x,s]),tabValues:s}}var x=t(6425);const f={tabList:"tabList_m0Yb",tabItem:"tabItem_Mhwx"};var g=t(7557);function T(e){let{className:n,block:t,selectedValue:a,selectValue:l,tabValues:i}=e;const o=[],{blockElementScrollPositionUntilNextRender:c}=(0,s.a_)(),u=e=>{const n=e.currentTarget,t=o.indexOf(n),r=i[t].value;r!==a&&(c(n),l(r))},p=e=>{let n=null;switch(e.key){case"Enter":u(e);break;case"ArrowRight":{const t=o.indexOf(e.currentTarget)+1;n=o[t]??o[0];break}case"ArrowLeft":{const t=o.indexOf(e.currentTarget)-1;n=o[t]??o[o.length-1];break}}n?.focus()};return(0,g.jsx)("ul",{role:"tablist","aria-orientation":"horizontal",className:(0,r.A)("tabs",{"tabs--block":t},n),children:i.map((e=>{let{value:n,label:t,attributes:s}=e;return(0,g.jsx)("li",{role:"tab",tabIndex:a===n?0:-1,"aria-selected":a===n,ref:e=>{o.push(e)},onKeyDown:p,onClick:u,...s,className:(0,r.A)("tabs__item",f.tabItem,s?.className,{"tabs__item--active":a===n}),children:t??n},n)}))})}function b(e){let{lazy:n,children:t,selectedValue:s}=e;const l=(Array.isArray(t)?t:[t]).filter(Boolean);if(n){const e=l.find((e=>e.props.value===s));return e?(0,a.cloneElement)(e,{className:(0,r.A)("margin-top--md",e.props.className)}):null}return(0,g.jsx)("div",{className:"margin-top--md",children:l.map(((e,n)=>(0,a.cloneElement)(e,{key:n,hidden:e.props.value!==s})))})}function j(e){const n=m(e);return(0,g.jsxs)("div",{className:(0,r.A)("tabs-container",f.tabList),children:[(0,g.jsx)(T,{...n,...e}),(0,g.jsx)(b,{...n,...e})]})}function v(e){const n=(0,x.A)();return(0,g.jsx)(j,{...e,children:p(e.children)},String(n))}}}]);