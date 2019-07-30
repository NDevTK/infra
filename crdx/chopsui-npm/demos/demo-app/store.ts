
// Do *not* try to use protos in Redux state objects.  They don't deserialize back
// into protos, just into plain js objects.

declare global {
  interface Window {
    process?: Object;
    __REDUX_DEVTOOLS_EXTENSION_COMPOSE__?: typeof compose;
  }
}

import {compose, createStore, AnyAction} from 'redux';

// Application state types
export interface Todo {
  text: string;
  completed: boolean;
}

export interface User {
  name: string;
  id: number;
}

type VisibilityFilter = 'SHOW_COMPLETED' | 'SHOW_ALL';

export interface AppState {
  todos?: Todo[];
  visibilityFilter?: VisibilityFilter;
  user?: User;
}

// Action types:

interface AddTodoAction extends AnyAction {
  type: 'ADD_TODO';
  text: string;
}

interface ToggleTodoAction extends AnyAction {
  type: 'TOGGLE_TODO';
  index: number;
}

export function toggleTodo(index: number): ToggleTodoAction {
  return {
    type: 'TOGGLE_TODO',
    index: index,
  };
}

interface SetVisibilityFilterAction extends AnyAction {
  type: 'SET_VISIBILITY_FILTER'
  filter: VisibilityFilter
}

interface UpdateUserAction extends AnyAction {
  type: 'UPDATE_USER'
  user: User
}

export function updateUser(user: User): UpdateUserAction {
  return {
    type: 'UPDATE_USER',
    user: user,
  }
}

// A "discriminated union" type to combine our action types.
export type TodoAppAction = AddTodoAction | ToggleTodoAction | SetVisibilityFilterAction | UpdateUserAction;

// type Reducer<S> = (state: S, action: AnyAction) => S;
// Reducer types:

function visibilityFilter(state: VisibilityFilter = 'SHOW_ALL', action: TodoAppAction): VisibilityFilter {
  if (action.type === 'SET_VISIBILITY_FILTER') {
      return action.filter
  } else {
      return state
  }
}
​
function todos(state: Todo[] = [{text: 'foo', completed:false}], action: TodoAppAction): Todo[] {
  switch (action.type) {
      case 'ADD_TODO':
          return state.concat([{text: action.text, completed: false}]);
      case 'TOGGLE_TODO':
          return state.map(
              (todo, index) =>
                  action.index === index
                      ? {text: todo.text, completed: !todo.completed}
                      : todo
          );
      default:
          return state;
  }
}

function user(state: User = {name: '', id: -1}, action: TodoAppAction): User {
  switch(action.type) {
       case 'UPDATE_USER':
            return Object.assign({}, state, action.user);
      default:
          return state;
  }
}

// This is our 'root' reducer:
function todoApp(state: AppState = {}, action: TodoAppAction): AppState {
  return {
      todos: todos(state.todos, action),
      visibilityFilter: visibilityFilter(state.visibilityFilter, action),
      user: user(state.user, action)
  }
}

// Store

// For debugging with the Redux Devtools extension:
// https://chrome.google.com/webstore/detail/redux-devtools/lmhkpmbekcpmknklioeibfkpmmfibljd/
//const composeEnhancers = window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__ || compose;


// This doesn't pass the linter:
export const store = createStore<AppState, TodoAppAction, any, any>(todoApp);

// This doesn't pass the compiler:
//export const store = createStore(todoApp);
// Compiler wins!
// TODO: figure out what to use instead of any, any. Or how to get naked
// createStore() to work with the compiler.
// https://github.com/reduxjs/redux/blob/master/index.d.ts#L305
