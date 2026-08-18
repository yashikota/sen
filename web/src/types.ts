export type IssueStatus = 'backlog' | 'todo' | 'in_progress' | 'done' | 'canceled';

export type Label = {
  id: number;
  name: string;
  color: string;
};

export type Issue = {
  id: number;
  number: number;
  identifier: string;
  title: string;
  body: string;
  status: IssueStatus;
  priority: number;
  projectId: number | null;
  projectSlug?: string | null;
  cycleId: number | null;
  cycleNumber?: number | null;
  dueDate: string | null;
  sortOrder: number;
  labels: Label[];
  createdAt: string;
  updatedAt: string;
  completedAt: string | null;
};

export type Project = {
  id: number;
  name: string;
  slug: string;
  description: string;
  status: string;
  startDate: string | null;
  targetDate: string | null;
  progress: number;
  createdAt: string;
  updatedAt: string;
};

export type Cycle = {
  id: number;
  number: number;
  startsAt: string;
  endsAt: string;
  status: string;
  createdAt: string;
  updatedAt: string;
};

export type Page = {
  id: number;
  title: string;
  slug: string;
  body: string;
  parentId: number | null;
  projectId: number | null;
  status: string;
  date: string | null;
  tags: string[];
  createdAt: string;
  updatedAt: string;
};

export type Comment = {
  id: number;
  issueId: number;
  body: string;
  createdAt: string;
};

export type Activity = {
  id: number;
  entityType: string;
  entityId: number;
  action: string;
  payload: Record<string, unknown>;
  createdAt: string;
};

export type Workspace = {
  name: string;
  ghcrRef: string;
  timezone: string;
  lastPushedAt: string | null;
  lastPushedDigest: string | null;
  updatedAt: string;
};

export type SearchHit = {
  kind: string;
  id: string;
  title: string;
};

export const ISSUE_STATUSES: IssueStatus[] = ['backlog', 'todo', 'in_progress', 'done', 'canceled'];

export const STATUS_LABEL: Record<IssueStatus, string> = {
  backlog: 'Backlog',
  todo: 'Todo',
  in_progress: 'In Progress',
  done: 'Done',
  canceled: 'Canceled',
};

export const PRIORITY_LABEL = ['No priority', 'Urgent', 'High', 'Medium', 'Low'];
