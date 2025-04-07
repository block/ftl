"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[744],{7389:(e,t,s)=>{s.d(t,{R:()=>c,x:()=>i});var n=s(8225);const l={},r=n.createContext(l);function c(e){const t=n.useContext(r);return n.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function i(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(l):e.components||l:c(e.components),n.createElement(r.Provider,{value:t},e.children)}},8649:(e,t,s)=>{s.r(t),s.d(t,{assets:()=>a,contentTitle:()=>i,default:()=>u,frontMatter:()=>c,metadata:()=>n,toc:()=>o});const n=JSON.parse('{"id":"reference/unittests","title":"Unit Tests","description":"Build unit tests for your modules","source":"@site/docs/reference/unittests.md","sourceDirName":"reference","slug":"/reference/unittests","permalink":"/ftl/docs/reference/unittests","draft":false,"unlisted":false,"editUrl":"https://github.com/block/ftl/tree/main/docs/docs/reference/unittests.md","tags":[],"version":"current","sidebarPosition":10,"frontMatter":{"sidebar_position":10,"title":"Unit Tests","description":"Build unit tests for your modules"},"sidebar":"tutorialSidebar","previous":{"title":"Fixtures","permalink":"/ftl/docs/reference/fixtures"},"next":{"title":"PubSub","permalink":"/ftl/docs/reference/pubsub"}}');var l=s(7557),r=s(7389);const c={sidebar_position:10,title:"Unit Tests",description:"Build unit tests for your modules"},i="Unit Tests",a={},o=[{value:"Create a context",id:"create-a-context",level:2},{value:"Customization",id:"customization",level:2},{value:"Project files, configs and secrets",id:"project-files-configs-and-secrets",level:3},{value:"Databases",id:"databases",level:3},{value:"Maps",id:"maps",level:3},{value:"Calls",id:"calls",level:3},{value:"PubSub",id:"pubsub",level:3}];function d(e){const t={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",header:"header",li:"li",p:"p",pre:"pre",ul:"ul",...(0,r.R)(),...e.components};return(0,l.jsxs)(l.Fragment,{children:[(0,l.jsx)(t.header,{children:(0,l.jsx)(t.h1,{id:"unit-tests",children:"Unit Tests"})}),"\n",(0,l.jsx)(t.h2,{id:"create-a-context",children:"Create a context"}),"\n",(0,l.jsx)(t.p,{children:"When writing a unit test, first create a context:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"func ExampleTest(t *testing.Test) {\n    ctx := ftltest.Context(\n        // options go here\n    )\n}\n"})}),"\n",(0,l.jsxs)(t.p,{children:["FTL will help isolate what you want to test by restricting access to FTL features by default. You can expand what is available to test by adding options to ",(0,l.jsx)(t.code,{children:"ftltest.Context(...)"}),"."]}),"\n",(0,l.jsx)(t.p,{children:"In this default set up, FTL does the following:"}),"\n",(0,l.jsxs)(t.ul,{children:["\n",(0,l.jsxs)(t.li,{children:["prevents access to ",(0,l.jsx)(t.code,{children:"ftl.ConfigValue"})," and ",(0,l.jsx)(t.code,{children:"ftl.SecretValue"})," (",(0,l.jsx)(t.a,{href:"#project-files-configs-and-secrets",children:"See options"}),")"]}),"\n",(0,l.jsxs)(t.li,{children:["prevents access to ",(0,l.jsx)(t.code,{children:"ftl.Database"})," (",(0,l.jsx)(t.a,{href:"#databases",children:"See options"}),")"]}),"\n",(0,l.jsxs)(t.li,{children:["prevents access to ",(0,l.jsx)(t.code,{children:"ftl.MapHandle"})," (",(0,l.jsx)(t.a,{href:"#maps",children:"See options"}),")"]}),"\n",(0,l.jsxs)(t.li,{children:["disables all subscribers (",(0,l.jsx)(t.a,{href:"#pubsub",children:"See options"}),")"]}),"\n"]}),"\n",(0,l.jsx)(t.h2,{id:"customization",children:"Customization"}),"\n",(0,l.jsx)(t.h3,{id:"project-files-configs-and-secrets",children:"Project files, configs and secrets"}),"\n",(0,l.jsx)(t.p,{children:"To enable configs and secrets from the default project file:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"ctx := ftltest.Context(\n    ftltest.WithDefaultProjectFile(),\n)\n"})}),"\n",(0,l.jsx)(t.p,{children:"Or you can specify a specific project file:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"ctx := ftltest.Context(\n    ftltest.WithProjectFile(path),\n)\n"})}),"\n",(0,l.jsx)(t.p,{children:"You can also override specific config and secret values:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:'ctx := ftltest.Context(\n    ftltest.WithDefaultProjectFile(),\n    \n    ftltest.WithConfig(endpoint, "test"),\n    ftltest.WithSecret(secret, "..."),\n)\n'})}),"\n",(0,l.jsx)(t.h3,{id:"databases",children:"Databases"}),"\n",(0,l.jsxs)(t.p,{children:["To enable database access in a test, you must first ",(0,l.jsx)(t.a,{href:"#project-files-configs-and-secrets",children:"provide a DSN via a project file"}),". You can then set up a test database:"]}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"ctx := ftltest.Context(\n    ftltest.WithDefaultProjectFile(),\n    ftltest.WithDatabase[MyDBConfig](),\n)\n"})}),"\n",(0,l.jsx)(t.p,{children:"This will:"}),"\n",(0,l.jsxs)(t.ul,{children:["\n",(0,l.jsxs)(t.li,{children:["Take the provided DSN and appends ",(0,l.jsx)(t.code,{children:"_test"})," to the database name. Eg: ",(0,l.jsx)(t.code,{children:"accounts"})," becomes ",(0,l.jsx)(t.code,{children:"accounts_test"})]}),"\n",(0,l.jsx)(t.li,{children:"Wipe all tables in the database so each test run happens on a clean database"}),"\n"]}),"\n",(0,l.jsx)(t.p,{children:"You can access the database in your test using its handle:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"db, err := ftltest.GetDatabaseHandle[MyDBConfig]()\ndb.Get(ctx).Exec(...)\n"})}),"\n",(0,l.jsx)(t.h3,{id:"maps",children:"Maps"}),"\n",(0,l.jsxs)(t.p,{children:["By default, calling ",(0,l.jsx)(t.code,{children:"Get(ctx)"})," on a map handle will panic."]}),"\n",(0,l.jsx)(t.p,{children:"You can inject a fake via a map:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:'ctx := ftltest.Context(\n    ftltest.WhenMap(exampleMap, func(ctx context.Context) (string, error) {\n       return "Test Value"\n    }),\n)\n'})}),"\n",(0,l.jsx)(t.p,{children:"You can also allow the use of all maps:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"ctx := ftltest.Context(\n    ftltest.WithMapsAllowed(),\n)\n"})}),"\n",(0,l.jsx)(t.h3,{id:"calls",children:"Calls"}),"\n",(0,l.jsxs)(t.p,{children:["Use ",(0,l.jsx)(t.code,{children:"ftltest.Call[Client](...)"})," (or ",(0,l.jsx)(t.code,{children:"ftltest.CallSource[Client](...)"}),", ",(0,l.jsx)(t.code,{children:"ftltest.CallSink[Client](...)"}),", ",(0,l.jsx)(t.code,{children:"ftltest.CallEmpty[Client](...)"}),") to invoke your\nverb. At runtime, FTL automatically provides these resources to your verb, and using ",(0,l.jsx)(t.code,{children:"ftltest.Call(...)"})," rather than direct invocations simulates this behavior."]}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:'// Call a verb\nresp, err := ftltest.Call[ExampleVerbClient, Request, Response](ctx, Request{Param: "Test"})\n'})}),"\n",(0,l.jsx)(t.p,{children:"You can inject fakes for verbs:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:'ctx := ftltest.Context(\n    ftltest.WhenVerb[ExampleVerbClient](func(ctx context.Context, req Request) (Response, error) {\n       return Response{Result: "Lorem Ipsum"}, nil\n    }),\n)\n'})}),"\n",(0,l.jsxs)(t.p,{children:["If there is no request or response parameters, you can use ",(0,l.jsx)(t.code,{children:"WhenSource(...)"}),", ",(0,l.jsx)(t.code,{children:"WhenSink(...)"}),", or ",(0,l.jsx)(t.code,{children:"WhenEmpty(...)"}),"."]}),"\n",(0,l.jsx)(t.p,{children:"To enable all calls within a module:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"ctx := ftltest.Context(\n    ftltest.WithCallsAllowedWithinModule(),\n)\n"})}),"\n",(0,l.jsx)(t.h3,{id:"pubsub",children:"PubSub"}),"\n",(0,l.jsx)(t.p,{children:"By default, all subscribers are disabled.\nTo enable a subscriber:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"ctx := ftltest.Context(\n    ftltest.WithSubscriber(paymentsSubscription, ProcessPayment),\n)\n"})}),"\n",(0,l.jsx)(t.p,{children:"Or you can inject a fake subscriber:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:'ctx := ftltest.Context(\n    ftltest.WithSubscriber(paymentsSubscription, func (ctx context.Context, in PaymentEvent) error {\n       return fmt.Errorf("failed payment: %v", in)\n    }),\n)\n'})}),"\n",(0,l.jsx)(t.p,{children:"Due to the asynchronous nature of pubsub, your test should wait for subscriptions to consume the published events:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:'topic.Publish(ctx, Event{Name: "Test"})\n\nftltest.WaitForSubscriptionsToComplete(ctx)\n// Event will have been consumed by now\n'})}),"\n",(0,l.jsx)(t.p,{children:"You can check what events were published to a topic:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"events := ftltest.EventsForTopic(ctx, topic)\n"})}),"\n",(0,l.jsx)(t.p,{children:"You can check what events were consumed by a subscription, and whether a subscriber returned an error:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"results := ftltest.ResultsForSubscription(ctx, subscription)\n"})}),"\n",(0,l.jsx)(t.p,{children:"If all you wanted to check was whether a subscriber returned an error, this function is simpler:"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-go",children:"errs := ftltest.ErrorsForSubscription(ctx, subscription)\n"})}),"\n",(0,l.jsx)(t.p,{children:"PubSub also has these different behaviours while testing:"}),"\n",(0,l.jsxs)(t.ul,{children:["\n",(0,l.jsx)(t.li,{children:"Publishing to topics in other modules is allowed"}),"\n",(0,l.jsx)(t.li,{children:"If a subscriber returns an error, no retries will occur regardless of retry policy."}),"\n"]})]})}function u(e={}){const{wrapper:t}={...(0,r.R)(),...e.components};return t?(0,l.jsx)(t,{...e,children:(0,l.jsx)(d,{...e})}):d(e)}}}]);