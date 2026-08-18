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
import { api } from './api.ts';

function NotFoundPage() {
  return <div className="empty">Not found</div>;
}

function ErrorPage({ error }: { error: Error }) {
  return <div className="error">{error.message}</div>;
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
    throw redirect({ to: '/issues' });
  },
  component: () => <Outlet />,
});

const issuesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/issues',
  loader: () => api.issues(),
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
  loader: () => api.issues(),
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
