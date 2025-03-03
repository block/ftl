"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[138],{1830:(e,t,r)=>{r.r(t),r.d(t,{default:()=>F});var s=r(8225),a=r(9112),n=r(57),c=r(6462),l=r(4288),o=r(5244),u=r(7207),i=r(3372),h=r(1654),d=r(4185),m=r(4344);const g=function(){const e=(0,d.A)(),t=(0,h.W6)(),r=(0,h.zy)(),{siteConfig:{baseUrl:s}}=(0,a.A)(),n=e?new URLSearchParams(r.search):null,c=n?.get("q")||"",l=n?.get("ctx")||"",o=n?.get("version")||"",u=e=>{const t=new URLSearchParams(r.search);return e?t.set("q",e):t.delete("q"),t};return{searchValue:c,searchContext:l&&Array.isArray(m.Hg)&&m.Hg.some((e=>"string"==typeof e?e===l:e.path===l))?l:"",searchVersion:o,updateSearchPath:e=>{const r=u(e);t.replace({search:r.toString()})},updateSearchContext:e=>{const s=new URLSearchParams(r.search);s.set("ctx",e),t.replace({search:s.toString()})},generateSearchPageLink:e=>{const t=u(e);return`${s}search?${t.toString()}`}}};var p=r(312),f=r(567),x=r(551),y=r(9207),j=r(6692),S=r(4303),w=r(5541);const A="searchContextInput_g1ux",C="searchQueryInput_Iqxb",v="searchResultItem_UcU7",P="searchResultItemPath_HNIH",b="searchResultItemSummary_B6RJ",R="searchQueryColumn_I7GL",T="searchContextColumn_z88a";var _=r(2331),$=r(7557);function H(){const{siteConfig:{baseUrl:e},i18n:{currentLocale:t}}=(0,a.A)(),{selectMessage:r}=(0,u.W)(),{searchValue:n,searchContext:l,searchVersion:h,updateSearchPath:d,updateSearchContext:f}=g(),[x,y]=(0,s.useState)(n),[j,w]=(0,s.useState)(),v=`${e}${h}`,P=(0,s.useMemo)((()=>x?(0,o.T)({id:"theme.SearchPage.existingResultsTitle",message:'Search results for "{query}"',description:"The search page title for non-empty query"},{query:x}):(0,o.T)({id:"theme.SearchPage.emptyResultsTitle",message:"Search the documentation",description:"The search page title for empty query"})),[x]);(0,s.useEffect)((()=>{d(x),x?(async()=>{const e=await(0,p.w)(v,l,x,100);w(e)})():w(void 0)}),[x,v,l]);const b=(0,s.useCallback)((e=>{y(e.target.value)}),[]);(0,s.useEffect)((()=>{n&&n!==x&&y(n)}),[n]);const[H,F]=(0,s.useState)(!1);return(0,s.useEffect)((()=>{!async function(){(!Array.isArray(m.Hg)||l||m.dz)&&await(0,p.k)(v,l),F(!0)}()}),[l,v]),(0,$.jsxs)(s.Fragment,{children:[(0,$.jsxs)(c.A,{children:[(0,$.jsx)("meta",{property:"robots",content:"noindex, follow"}),(0,$.jsx)("title",{children:P})]}),(0,$.jsxs)("div",{className:"container margin-vert--lg",children:[(0,$.jsx)("h1",{children:P}),(0,$.jsxs)("div",{className:"row",children:[(0,$.jsx)("div",{className:(0,i.A)("col",{[R]:Array.isArray(m.Hg),"col--9":Array.isArray(m.Hg),"col--12":!Array.isArray(m.Hg)}),children:(0,$.jsx)("input",{type:"search",name:"q",className:C,"aria-label":"Search",onChange:b,value:x,autoComplete:"off",autoFocus:!0})}),Array.isArray(m.Hg)?(0,$.jsx)("div",{className:(0,i.A)("col","col--3","padding-left--none",T),children:(0,$.jsxs)("select",{name:"search-context",className:A,id:"context-selector",value:l,onChange:e=>f(e.target.value),children:[m.dz&&(0,$.jsx)("option",{value:"",children:(0,o.T)({id:"theme.SearchPage.searchContext.everywhere",message:"Everywhere"})}),m.Hg.map((e=>{const{label:r,path:s}=(0,_.p)(e,t);return(0,$.jsx)("option",{value:s,children:r},s)}))]})}):null]}),!H&&x&&(0,$.jsx)("div",{children:(0,$.jsx)(S.A,{})}),j&&(j.length>0?(0,$.jsx)("p",{children:r(j.length,(0,o.T)({id:"theme.SearchPage.documentsFound.plurals",message:"1 document found|{count} documents found",description:'Pluralized label for "{count} documents found". Use as much plural forms (separated by "|") as your language support (see https://www.unicode.org/cldr/cldr-aux/charts/34/supplemental/language_plural_rules.html)'},{count:j.length}))}):(0,$.jsx)("p",{children:(0,o.T)({id:"theme.SearchPage.noResultsText",message:"No documents were found",description:"The paragraph for empty search result"})})),(0,$.jsx)("section",{children:j&&j.map((e=>(0,$.jsx)(I,{searchResult:e},e.document.i)))})]})]})}function I(e){let{searchResult:{document:t,type:r,page:s,tokens:a,metadata:n}}=e;const c=r===f.i.Title,o=r===f.i.Keywords,u=r===f.i.Description,i=u||o,h=c||i,d=r===f.i.Content,g=(c?t.b:s.b).slice(),p=d||i?t.s:t.t;h||g.push(s.t);let S="";if(m.CU&&a.length>0){const e=new URLSearchParams;for(const t of a)e.append("_highlight",t);S=`?${e.toString()}`}return(0,$.jsxs)("article",{className:v,children:[(0,$.jsx)("h2",{children:(0,$.jsx)(l.A,{to:t.u+S+(t.h||""),dangerouslySetInnerHTML:{__html:d||i?(0,x.Z)(p,a):(0,y.C)(p,(0,j.g)(n,"t"),a,100)}})}),g.length>0&&(0,$.jsx)("p",{className:P,children:(0,w.$)(g)}),(d||u)&&(0,$.jsx)("p",{className:b,dangerouslySetInnerHTML:{__html:(0,y.C)(t.t,(0,j.g)(n,"t"),a,100)}})]})}const F=function(){return(0,$.jsx)(n.A,{children:(0,$.jsx)(H,{})})}},7207:(e,t,r)=>{r.d(t,{W:()=>u});var s=r(8225),a=r(9112);const n=["zero","one","two","few","many","other"];function c(e){return n.filter((t=>e.includes(t)))}const l={locale:"en",pluralForms:c(["one","other"]),select:e=>1===e?"one":"other"};function o(){const{i18n:{currentLocale:e}}=(0,a.A)();return(0,s.useMemo)((()=>{try{return function(e){const t=new Intl.PluralRules(e);return{locale:e,pluralForms:c(t.resolvedOptions().pluralCategories),select:e=>t.select(e)}}(e)}catch(t){return console.error(`Failed to use Intl.PluralRules for locale "${e}".\nDocusaurus will fallback to the default (English) implementation.\nError: ${t.message}\n`),l}}),[e])}function u(){const e=o();return{selectMessage:(t,r)=>function(e,t,r){const s=e.split("|");if(1===s.length)return s[0];s.length>r.pluralForms.length&&console.error(`For locale=${r.locale}, a maximum of ${r.pluralForms.length} plural forms are expected (${r.pluralForms.join(",")}), but the message contains ${s.length}: ${e}`);const a=r.select(t),n=r.pluralForms.indexOf(a);return s[Math.min(n,s.length-1)]}(r,t,e)}}}}]);