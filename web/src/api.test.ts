import { describe, expect, it } from 'vite-plus/test';
import { issuesQuery, parseIssueSearch, searchToFilter } from './api.ts';

describe('issuesQuery', () => {
  it('returns empty string when no filters are set', () => {
    expect(issuesQuery({})).toBe('');
    expect(issuesQuery({ status: null, labels: [], priority: -1 })).toBe('');
  });

  it('encodes compound filters used by saved views', () => {
    expect(
      issuesQuery({
        status: 'todo',
        project: 'harbor',
        cycle: 2,
        labels: ['Bug', 'Feature'],
        priority: 1,
      }),
    ).toBe('?status=todo&project=harbor&cycle=2&labels=Bug%2CFeature&priority=1');
  });
});

describe('parseIssueSearch', () => {
  it('drops empty and invalid values', () => {
    expect(parseIssueSearch({ status: '', cycle: 'nope', priority: '' })).toEqual({});
  });

  it('accepts numeric cycle and priority zero', () => {
    expect(parseIssueSearch({ cycle: 3, priority: 0 })).toEqual({ cycle: 3, priority: 0 });
  });

  it('drops cycle zero', () => {
    expect(parseIssueSearch({ cycle: 0 })).toEqual({});
  });

  it('keeps the same fields saved views use', () => {
    expect(
      parseIssueSearch({
        status: 'todo',
        project: 'harbor',
        cycle: '2',
        priority: '1',
        labels: 'Bug,Feature',
      }),
    ).toEqual({
      status: 'todo',
      project: 'harbor',
      cycle: 2,
      priority: 1,
      labels: 'Bug,Feature',
    });
  });
});

describe('searchToFilter', () => {
  it('splits label names for the issues API', () => {
    expect(
      searchToFilter({
        status: 'todo',
        labels: 'Bug, Feature',
        priority: 1,
      }),
    ).toEqual({
      status: 'todo',
      project: undefined,
      cycle: undefined,
      labels: ['Bug', 'Feature'],
      priority: 1,
    });
  });

  it('omits labels when the search has none', () => {
    expect(searchToFilter({})).toEqual({
      status: undefined,
      project: undefined,
      cycle: undefined,
      labels: undefined,
      priority: undefined,
    });
  });
});

describe('issuesQuery priority zero', () => {
  it('encodes no-priority as priority=0', () => {
    expect(issuesQuery({ priority: 0 })).toBe('?priority=0');
  });
});
