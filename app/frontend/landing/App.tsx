import React from "react";
import { BrowserRouter, Routes, Route, Outlet } from "react-router-dom";
import { GlobalContextProviders } from "./components/_globalContextProviders";
import Page_0 from "./pages/_index.tsx";
import PageLayout_0 from "./pages/_index.pageLayout.tsx";
import Page_Privacy from "./pages/privacy.tsx";
import Terms_of_Service from "./pages/terms.tsx";
import Contact_Page from "./pages/contact.tsx";
import Integrations_Page from "./pages/integrations.tsx";

if (!window.requestIdleCallback) {
  window.requestIdleCallback = (cb) => {
    const id = setTimeout(cb, 1);
    // Type of setTimeout in browsers is number
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    return id as number;
  };
}

import "./base.css";

const fileNameToRoute = new Map([
  ["./pages/_index.tsx", "/"],
  ["./pages/privacy.tsx", "/privacy"],
  ["./pages/terms.tsx", "/terms"],
  ["./pages/contact.tsx", "/contact"],
  ["./pages/integrations.tsx", "/integrations"],
  
]);
const fileNameToComponent: Map<string, React.ComponentType<any>> = new Map([
  ["./pages/_index.tsx", Page_0],
  ["./pages/privacy.tsx", Page_Privacy],
  ["./pages/terms.tsx", Terms_of_Service],
  ["./pages/contact.tsx", Contact_Page],
  ["./pages/integrations.tsx", Integrations_Page],
]);

function makePageRoute(filename: string) {
  const Component = fileNameToComponent.get(filename);
  if (!Component) {
    // If a page component isn't registered, render nothing (router will handle unknown paths),
    // or you can return a NotFound element instead.
    return null;
  }
  const C = Component as React.ComponentType<any>;
  return <C />;
}

function toElement({
  trie,
  fileNameToRoute,
  makePageRoute,
}: {
  trie: LayoutTrie;
  fileNameToRoute: Map<string, string>;
  makePageRoute: (filename: string) => React.ReactNode;
}) {
  return [
    ...trie.topLevel.map((filename) => (
      <Route
        key={fileNameToRoute.get(filename)}
        path={fileNameToRoute.get(filename)}
        element={makePageRoute(filename)}
      />
    )),
    ...Array.from(trie.trie.entries()).map(([Component, child], index) => (
      <Route
        key={index}
        element={
          <Component>
            <Outlet />
          </Component>
        }
      >
        {toElement({ trie: child, fileNameToRoute, makePageRoute })}
      </Route>
    )),
  ];
}

type LayoutTrieNode = Map<
  React.ComponentType<{ children: React.ReactNode }>,
  LayoutTrie
>;
type LayoutTrie = { topLevel: string[]; trie: LayoutTrieNode };
function buildLayoutTrie(layouts: {
  [fileName: string]: React.ComponentType<{ children: React.ReactNode }>[];
}): LayoutTrie {
  const result: LayoutTrie = { topLevel: [], trie: new Map() };
  Object.entries(layouts).forEach(([fileName, components]) => {
    let cur: LayoutTrie = result;
    for (const component of components) {
      if (!cur.trie.has(component)) {
        cur.trie.set(component, {
          topLevel: [],
          trie: new Map(),
        });
      }
      cur = cur.trie.get(component)!;
    }
    cur.topLevel.push(fileName);
  });
  return result;
}

function NotFound() {
  return (
    <div>
      <h1>Not Found</h1>
      <p>The page you are looking for does not exist.</p>
      <p>Go back to the <a href="/" style={{ color: 'blue' }}>home page</a>.</p>
    </div>
  );
}

export function App() {
  return (
    <BrowserRouter>
      <GlobalContextProviders>
        <Routes>
          {toElement({ trie: buildLayoutTrie({
"./pages/_index.tsx": PageLayout_0,
"./pages/privacy.tsx": PageLayout_0,
"./pages/terms.tsx": PageLayout_0,
"./pages/contact.tsx": PageLayout_0,
"./pages/integrations.tsx": PageLayout_0,
}), fileNameToRoute, makePageRoute })} 
          <Route path="*" element={<NotFound />} />
        </Routes>
      </GlobalContextProviders>
    </BrowserRouter>
  );
}
