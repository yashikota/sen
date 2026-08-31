import {
  Outlet,
  createRootRoute,
  createRoute,
  createRouter,
  redirect,
} from '@tanstack/react-router';
import { Shell } from './components/Shell.tsx';
import { BoardPage, IssueRoutePage, IssuesPage } from './pages/IssuesPages.tsx';
import { PageDetailPage, PagesPage } from './pages/PagesPages.tsx';
import {
  CycleDetailPage,
  CyclesPage,
  ProjectDetailPage,
  ProjectsPage,
} from './pages/ProjectsCycles.tsx';
import { ViewPage } from './pages/ViewsPages.tsx';
import { api, issuesQuery, parseIssueSearch, searchToFilter, type IssueSearch } from './api.ts';

function NotFoundPage() {
  return <div className="empty">Not found</div>;
}

function ErrorPage({ error }: { error: Error }) {
  return <div className="error">{error.message}</div>;
}

async function loadFilteredIssues(search: IssueSearch) {
  const [issues, projects, cycles, labels] = await Promise.all([
    api.issues(issuesQuery(searchToFilter(search))),
    api.projects(),
    api.cycles(),
    api.labels(),
  ]);
  return { issues: issues ?? [], projects, cycles, labels };
}

const rootRoute = createRootRoute({
  component: Shell,
  notFoundComponent: NotFoundPage,
  errorComponent: ErrorPage,
});

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  beforeLoad: () => {
    throw redirect({ to: '/issues', search: {} });
  },
  component: () => <Outlet />,
});

const issuesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/issues',
  validateSearch: (raw: Record<string, unknown>) => parseIssueSearch(raw),
  loaderDeps: ({ search }) => search,
  loader: ({ deps }) => loadFilteredIssues(deps),
  component: IssuesPage,
});

const issueRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/issues/$identifier',
  loader: () => api.issues(),
  component: IssueRoutePage,
});

const boardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/board',
  validateSearch: (raw: Record<string, unknown>) => parseIssueSearch(raw),
  loaderDeps: ({ search }) => search,
  loader: ({ deps }) => loadFilteredIssues(deps),
  component: BoardPage,
});

const projectsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/projects',
  loader: () => api.projects(),
  component: ProjectsPage,
});

const projectRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/projects/$slug',
  loader: async ({ params }) => ({
    project: await api.project(params.slug),
    issues: await api.issues(`?project=${encodeURIComponent(params.slug)}`),
  }),
  component: ProjectDetailPage,
});

const cyclesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/cycles',
  loader: () => api.cycles(),
  component: CyclesPage,
});

const cycleRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/cycles/$number',
  loader: async ({ params }) => {
    const number = Number(params.number);
    return {
      cycle: await api.cycle(number),
      issues: await api.issues(`?cycle=${number}`),
    };
  },
  component: CycleDetailPage,
});

const viewRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/views/$slug',
  loader: async ({ params }) => {
    const view = await api.view(params.slug);
    const [issues, projects, cycles, labels] = await Promise.all([
      api.issues(issuesQuery(view)),
      api.projects(),
      api.cycles(),
      api.labels(),
    ]);
    return { view, issues, projects, cycles, labels };
  },
  component: ViewPage,
});

const pagesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/pages',
  loader: () => api.pages(),
  component: PagesPage,
});

const pageRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/pages/$slug',
  loader: ({ params }) => api.page(params.slug),
  component: PageDetailPage,
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  issuesRoute,
  issueRoute,
  boardRoute,
  projectsRoute,
  projectRoute,
  cyclesRoute,
  cycleRoute,
  viewRoute,
  pagesRoute,
  pageRoute,
]);

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 30_000,
  scrollRestoration: true,
});

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
