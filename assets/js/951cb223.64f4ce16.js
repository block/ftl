"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[788],{78:(e,n,t)=>{t.d(n,{A:()=>a});t(8225);var i=t(3372);const o={tabItem:"tabItem_k9Yy"};var r=t(7557);function a(e){let{children:n,hidden:t,className:a}=e;return(0,r.jsx)("div",{role:"tabpanel",className:(0,i.A)(o.tabItem,a),hidden:t,children:n})}},232:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>u,contentTitle:()=>l,default:()=>h,frontMatter:()=>c,metadata:()=>i,toc:()=>p});const i=JSON.parse('{"id":"reference/pubsub","title":"PubSub","description":"Asynchronous publishing of events to topics","source":"@site/docs/reference/pubsub.md","sourceDirName":"reference","slug":"/reference/pubsub","permalink":"/ftl/docs/reference/pubsub","draft":false,"unlisted":false,"editUrl":"https://github.com/block/ftl/tree/main/docs/docs/reference/pubsub.md","tags":[],"version":"current","sidebarPosition":11,"frontMatter":{"sidebar_position":11,"title":"PubSub","description":"Asynchronous publishing of events to topics"},"sidebar":"tutorialSidebar","previous":{"title":"Unit Tests","permalink":"/ftl/docs/reference/unittests"},"next":{"title":"Retries","permalink":"/ftl/docs/reference/retries"}}');var o=t(7557),r=t(6039),a=t(1414),s=t(78);const c={sidebar_position:11,title:"PubSub",description:"Asynchronous publishing of events to topics"},l="PubSub",u={},p=[{value:"Declaring a Topic",id:"declaring-a-topic",level:2},{value:"Multi-Partition Topics",id:"multi-partition-topics",level:2},{value:"Publishing Events",id:"publishing-events",level:2},{value:"Subscribing to Topics",id:"subscribing-to-topics",level:2}];function d(e){const n={code:"code",h1:"h1",h2:"h2",header:"header",p:"p",pre:"pre",...(0,r.R)(),...e.components};return(0,o.jsxs)(o.Fragment,{children:[(0,o.jsx)(n.header,{children:(0,o.jsx)(n.h1,{id:"pubsub",children:"PubSub"})}),"\n",(0,o.jsx)(n.p,{children:"FTL has first-class support for PubSub, modelled on the concepts of topics (where events are sent) and subscribers (a verb which consumes events). Subscribers are, as you would expect, sinks. Each subscriber is a cursor over the topic it is associated with. Each topic may have multiple subscriptions. Each published event has an at least once delivery guarantee for each subscription."}),"\n",(0,o.jsx)(n.p,{children:"A topic can be exported to allow other modules to subscribe to it. Subscriptions are always private to their module."}),"\n",(0,o.jsx)(n.p,{children:"When a subscription is first created in an environment, it can start consuming from the beginning of the topic or only consume events published afterwards."}),"\n",(0,o.jsx)(n.p,{children:"Topics allow configuring the number of partitions and how each event should be mapped to a partition, allowing for greater throughput. Subscriptions will consume in order within each partition. There are cases where a small amount of progress on a subscription will be lost, so subscriptions should be able to handle receiving some events that have already been consumed."}),"\n","\n",(0,o.jsx)(n.h2,{id:"declaring-a-topic",children:"Declaring a Topic"}),"\n",(0,o.jsx)(n.p,{children:"Here's how to declare a simple topic with a single partition:"}),"\n",(0,o.jsxs)(a.A,{groupId:"languages",children:[(0,o.jsxs)(s.A,{value:"go",label:"Go",default:!0,children:[(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:'package payments\n\nimport (\n  "github.com/block/ftl/go-runtime/ftl"\n)\n\n// Define an event type\ntype Invoice struct {\n  InvoiceNo string\n}\n\n//ftl:topic partitions=1\ntype Invoices = ftl.TopicHandle[Invoice, ftl.SinglePartitionMap[Invoice]]\n'})}),(0,o.jsx)(n.p,{children:"Note that the name of the topic as represented in the FTL schema is the lower camel case version of the type name."}),(0,o.jsxs)(n.p,{children:["The ",(0,o.jsx)(n.code,{children:"Invoices"})," type is a handle to the topic. It is a generic type that takes two arguments: the event type and the partition map type. The partition map type is used to map events to partitions."]})]}),(0,o.jsx)(s.A,{value:"kotlin",label:"Kotlin",children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-kotlin",children:'import xyz.block.ftl.Export;\nimport xyz.block.ftl.SinglePartitionMapper\nimport xyz.block.ftl.Topic\nimport xyz.block.ftl.WriteableTopic\n\n// Define the event type for the topic\ndata class Invoice(val invoiceNo: String)\n\n// Add @Export if you want other modules to be able to consume from this topic\n@Topic(name = "invoices", partitions = 1)\ninternal interface InvoicesTopic : WriteableTopic<Invoice, SinglePartitionMapper>\n'})})}),(0,o.jsx)(s.A,{value:"java",label:"Java",children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-java",children:'import xyz.block.ftl.Export;\nimport xyz.block.ftl.SinglePartitionMapper;\nimport xyz.block.ftl.Topic;\nimport xyz.block.ftl.WriteableTopic;\n\n// Define the event type for the topic\nrecord Invoice(String invoiceNo) {\n}\n\n// Add @Export if you want other modules to be able to consume from this topic\n@Topic(name = "invoices", partitions = 1)\ninterface InvoicesTopic extends WriteableTopic<Invoice, SinglePartitionMapper> {\n}\n'})})}),(0,o.jsx)(s.A,{value:"schema",label:"Schema",children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-schema",children:"module payments {\n  // The Invoice data type that will be published to the topic\n  data Invoice {\n    invoiceNo String\n  }\n\n  // A topic with a single partition\n  topic invoices payments.Invoice\n}\n"})})})]}),"\n",(0,o.jsx)(n.h2,{id:"multi-partition-topics",children:"Multi-Partition Topics"}),"\n",(0,o.jsx)(n.p,{children:"For topics that require multiple partitions, you'll need to implement a partition mapper:"}),"\n",(0,o.jsxs)(a.A,{groupId:"languages",children:[(0,o.jsx)(s.A,{value:"go",label:"Go",default:!0,children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:'package payments\n\nimport (\n  "github.com/block/ftl/go-runtime/ftl"\n)\n\n// Define an event type\ntype Invoice struct {\n  InvoiceNo string\n}\n\ntype PartitionMapper struct{}\n\nvar _ ftl.TopicPartitionMap[PubSubEvent] = PartitionMapper{}\n\nfunc (PartitionMapper) PartitionKey(event PubSubEvent) string {\n\treturn event.Time.String()\n}\n\n//ftl:topic partitions=10\ntype Invoices = ftl.TopicHandle[Invoice, PartitionMapper]\n'})})}),(0,o.jsx)(s.A,{value:"kotlin",label:"Kotlin",children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-kotlin",children:'import xyz.block.ftl.Export;\nimport xyz.block.ftl.SinglePartitionMapper\nimport xyz.block.ftl.Topic\nimport xyz.block.ftl.TopicPartitionMapper\nimport xyz.block.ftl.WriteableTopic\n\n// Define the event type for the topic\ndata class Invoice(val invoiceNo: String)\n\n// PartitionMapper maps each to a partition in the topic\nclass PartitionMapper : TopicPartitionMapper<Invoice> {\n    override fun getPartitionKey(invoice: Invoice): String {\n        return invoice.invoiceNo\n    }\n}\n\n// Add @Export if you want other modules to be able to consume from this topic\n@Topic(name = "invoices", partitions = 8)\ninternal interface InvoicesTopic : WriteableTopic<Invoice, PartitionMapper>\n'})})}),(0,o.jsx)(s.A,{value:"java",label:"Java",children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-java",children:'import xyz.block.ftl.Export;\nimport xyz.block.ftl.Topic;\nimport xyz.block.ftl.TopicPartitionMapper;\nimport xyz.block.ftl.WriteableTopic;\n\n// Define the event type for the topic\nrecord Invoice(String invoiceNo) {\n}\n\n// PartitionMapper maps each to a partition in the topic\nclass PartitionMapper implements TopicPartitionMapper<Invoice> {\n    public String getPartitionKey(Invoice invoice) {\n        return invoice.invoiceNo();\n    }\n}\n\n// Add @Export if you want other modules to be able to consum from this topic\n@Topic(name = "invoices", partitions = 8)\ninterface InvoicesTopic extends WriteableTopic<Invoice, PartitionMapper> {\n}\n'})})}),(0,o.jsx)(s.A,{value:"schema",label:"Schema",children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-schema",children:"module payments {\n  // The Invoice data type that will be published to the topic\n  data Invoice {\n    invoiceNo String\n  }\n\n  // A topic with multiple partitions (8 or 10 depending on language)\n  // The partition key is determined by the mapper implementation\n  topic invoices payments.Invoice\n    +partitions 8\n}\n"})})})]}),"\n",(0,o.jsx)(n.h2,{id:"publishing-events",children:"Publishing Events"}),"\n",(0,o.jsx)(n.p,{children:"Events can be published to a topic by injecting the topic into a verb:"}),"\n",(0,o.jsxs)(a.A,{groupId:"languages",children:[(0,o.jsx)(s.A,{value:"go",label:"Go",default:!0,children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:"//ftl:verb\nfunc PublishInvoice(ctx context.Context, topic Invoices) error {\n   topic.Publish(ctx, Invoice{...})\n   // ...\n}\n"})})}),(0,o.jsx)(s.A,{value:"kotlin",label:"Kotlin",children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-kotlin",children:"@Verb\nfun publishInvoice(request: InvoiceRequest, topic: InvoicesTopic) {\n    topic.publish(Invoice(request.invoiceNo))\n}\n"})})}),(0,o.jsx)(s.A,{value:"java",label:"Java",children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-java",children:"@Verb\nvoid publishInvoice(InvoiceRequest request, InvoicesTopic topic) throws Exception {\n    topic.publish(new Invoice(request.invoiceNo()));\n}\n"})})}),(0,o.jsx)(s.A,{value:"schema",label:"Schema",children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-schema",children:"module payments {\n  data InvoiceRequest {\n    invoiceNo String\n  }\n  \n  data Invoice {\n    invoiceNo String\n  }\n  \n  topic invoices payments.Invoice\n  \n  // A verb that publishes to the invoices topic\n  verb publishInvoice(payments.InvoiceRequest) Unit\n    +publish payments.invoices\n}\n"})})})]}),"\n",(0,o.jsx)(n.h2,{id:"subscribing-to-topics",children:"Subscribing to Topics"}),"\n",(0,o.jsx)(n.p,{children:"Here's how to subscribe to topics:"}),"\n",(0,o.jsxs)(a.A,{groupId:"languages",children:[(0,o.jsx)(s.A,{value:"go",label:"Go",default:!0,children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:"// Configure initial event consumption with either from=beginning or from=latest\n//\n//ftl:subscribe payments.invoices from=beginning\nfunc SendInvoiceEmail(ctx context.Context, in Invoice) error {\n  // ...\n}\n"})})}),(0,o.jsxs)(s.A,{value:"kotlin",label:"Kotlin",children:[(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-kotlin",children:"// if subscribing from another module, import the event and topic\nimport ftl.publisher.Invoice\nimport ftl.publisher.InvoicesTopic\n\nimport xyz.block.ftl.FromOffset\nimport xyz.block.ftl.Subscription\n\n@Subscription(topic = InvoicesTopic::class, from = FromOffset.LATEST)\nfun consumeInvoice(event: Invoice) {\n    // ...\n}\n"})}),(0,o.jsx)(n.p,{children:"If you are subscribing to a topic from another module, FTL will generate a topic class for you so you can subscribe to it. This generated\ntopic cannot be published to, only subscribed to:"}),(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-kotlin",children:'@Topic(name="invoices", module="publisher")\ninternal interface InvoicesTopic : ConsumableTopic<Invoice>\n'})})]}),(0,o.jsxs)(s.A,{value:"java",label:"Java",children:[(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-java",children:"// if subscribing from another module, import the event and topic\nimport ftl.othermodule.Invoice;\nimport ftl.othermodule.InvoicesTopic;\n\nimport xyz.block.ftl.FromOffset;\nimport xyz.block.ftl.Subscription;\n\nclass Subscriber {\n    @Subscription(topic = InvoicesTopic.class, from = FromOffset.LATEST)\n    public void consumeInvoice(Invoice event) {\n        // ...\n    }\n}\n"})}),(0,o.jsx)(n.p,{children:"If you are subscribing to a topic from another module, FTL will generate a topic class for you so you can subscribe to it. This generated\ntopic cannot be published to, only subscribed to:"}),(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-java",children:'@Topic(name="invoices", module="publisher")\ninterface InvoicesTopic extends ConsumableTopic<Invoice> {}\n'})})]}),(0,o.jsx)(s.A,{value:"schema",label:"Schema",children:(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-schema",children:"module payments {\n  data InvoiceRequest {\n    invoiceNo String\n  }\n  \n  data Invoice {\n    invoiceNo String\n  }\n  \n  topic invoices payments.Invoice\n  \n  // A verb that subscribes to the invoices topic\n  verb sendInvoiceEmail(payments.Invoice) Unit\n    +subscribe payments.invoices from=beginning\n}\n\n// In another module\nmodule emailer {\n  // A verb that subscribes to the invoices topic from another module\n  verb consumeInvoice(payments.Invoice) Unit\n    +subscribe payments.invoices from=latest\n}\n"})})})]})]})}function h(e={}){const{wrapper:n}={...(0,r.R)(),...e.components};return n?(0,o.jsx)(n,{...e,children:(0,o.jsx)(d,{...e})}):d(e)}},1414:(e,n,t)=>{t.d(n,{A:()=>j});var i=t(8225),o=t(3372),r=t(2646),a=t(1654),s=t(8322),c=t(9467),l=t(1544),u=t(5126);function p(e){return i.Children.toArray(e).filter((e=>"\n"!==e)).map((e=>{if(!e||(0,i.isValidElement)(e)&&function(e){const{props:n}=e;return!!n&&"object"==typeof n&&"value"in n}(e))return e;throw new Error(`Docusaurus error: Bad <Tabs> child <${"string"==typeof e.type?e.type:e.type.name}>: all children of the <Tabs> component should be <TabItem>, and every <TabItem> should have a unique "value" prop.`)}))?.filter(Boolean)??[]}function d(e){const{values:n,children:t}=e;return(0,i.useMemo)((()=>{const e=n??function(e){return p(e).map((e=>{let{props:{value:n,label:t,attributes:i,default:o}}=e;return{value:n,label:t,attributes:i,default:o}}))}(t);return function(e){const n=(0,l.XI)(e,((e,n)=>e.value===n.value));if(n.length>0)throw new Error(`Docusaurus error: Duplicate values "${n.map((e=>e.value)).join(", ")}" found in <Tabs>. Every value needs to be unique.`)}(e),e}),[n,t])}function h(e){let{value:n,tabValues:t}=e;return t.some((e=>e.value===n))}function b(e){let{queryString:n=!1,groupId:t}=e;const o=(0,a.W6)(),r=function(e){let{queryString:n=!1,groupId:t}=e;if("string"==typeof n)return n;if(!1===n)return null;if(!0===n&&!t)throw new Error('Docusaurus error: The <Tabs> component groupId prop is required if queryString=true, because this value is used as the search param name. You can also provide an explicit value such as queryString="my-search-param".');return t??null}({queryString:n,groupId:t});return[(0,c.aZ)(r),(0,i.useCallback)((e=>{if(!r)return;const n=new URLSearchParams(o.location.search);n.set(r,e),o.replace({...o.location,search:n.toString()})}),[r,o])]}function m(e){const{defaultValue:n,queryString:t=!1,groupId:o}=e,r=d(e),[a,c]=(0,i.useState)((()=>function(e){let{defaultValue:n,tabValues:t}=e;if(0===t.length)throw new Error("Docusaurus error: the <Tabs> component requires at least one <TabItem> children component");if(n){if(!h({value:n,tabValues:t}))throw new Error(`Docusaurus error: The <Tabs> has a defaultValue "${n}" but none of its children has the corresponding value. Available values are: ${t.map((e=>e.value)).join(", ")}. If you intend to show no default tab, use defaultValue={null} instead.`);return n}const i=t.find((e=>e.default))??t[0];if(!i)throw new Error("Unexpected error: 0 tabValues");return i.value}({defaultValue:n,tabValues:r}))),[l,p]=b({queryString:t,groupId:o}),[m,v]=function(e){let{groupId:n}=e;const t=function(e){return e?`docusaurus.tab.${e}`:null}(n),[o,r]=(0,u.Dv)(t);return[o,(0,i.useCallback)((e=>{t&&r.set(e)}),[t,r])]}({groupId:o}),f=(()=>{const e=l??m;return h({value:e,tabValues:r})?e:null})();(0,s.A)((()=>{f&&c(f)}),[f]);return{selectedValue:a,selectValue:(0,i.useCallback)((e=>{if(!h({value:e,tabValues:r}))throw new Error(`Can't select invalid tab value=${e}`);c(e),p(e),v(e)}),[p,v,r]),tabValues:r}}var v=t(8442);const f={tabList:"tabList_CY6c",tabItem:"tabItem_IDbK"};var g=t(7557);function x(e){let{className:n,block:t,selectedValue:i,selectValue:a,tabValues:s}=e;const c=[],{blockElementScrollPositionUntilNextRender:l}=(0,r.a_)(),u=e=>{const n=e.currentTarget,t=c.indexOf(n),o=s[t].value;o!==i&&(l(n),a(o))},p=e=>{let n=null;switch(e.key){case"Enter":u(e);break;case"ArrowRight":{const t=c.indexOf(e.currentTarget)+1;n=c[t]??c[0];break}case"ArrowLeft":{const t=c.indexOf(e.currentTarget)-1;n=c[t]??c[c.length-1];break}}n?.focus()};return(0,g.jsx)("ul",{role:"tablist","aria-orientation":"horizontal",className:(0,o.A)("tabs",{"tabs--block":t},n),children:s.map((e=>{let{value:n,label:t,attributes:r}=e;return(0,g.jsx)("li",{role:"tab",tabIndex:i===n?0:-1,"aria-selected":i===n,ref:e=>{c.push(e)},onKeyDown:p,onClick:u,...r,className:(0,o.A)("tabs__item",f.tabItem,r?.className,{"tabs__item--active":i===n}),children:t??n},n)}))})}function y(e){let{lazy:n,children:t,selectedValue:r}=e;const a=(Array.isArray(t)?t:[t]).filter(Boolean);if(n){const e=a.find((e=>e.props.value===r));return e?(0,i.cloneElement)(e,{className:(0,o.A)("margin-top--md",e.props.className)}):null}return(0,g.jsx)("div",{className:"margin-top--md",children:a.map(((e,n)=>(0,i.cloneElement)(e,{key:n,hidden:e.props.value!==r})))})}function I(e){const n=m(e);return(0,g.jsxs)("div",{className:(0,o.A)("tabs-container",f.tabList),children:[(0,g.jsx)(x,{...n,...e}),(0,g.jsx)(y,{...n,...e})]})}function j(e){const n=(0,v.A)();return(0,g.jsx)(I,{...e,children:p(e.children)},String(n))}},6039:(e,n,t)=>{t.d(n,{R:()=>a,x:()=>s});var i=t(8225);const o={},r=i.createContext(o);function a(e){const n=i.useContext(r);return i.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function s(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(o):e.components||o:a(e.components),i.createElement(r.Provider,{value:n},e.children)}}}]);