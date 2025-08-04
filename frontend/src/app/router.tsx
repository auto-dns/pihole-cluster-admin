import { createBrowserRouter } from 'react-router';
// Routes
import {ProtectedRoute, ProtectedRouteFullInit} from './routes/ProtectedRoute';
import {UnprotectedRoute, UnprotectedRouteUninitialized} from './routes/UnprotectedRoute';
// Pages
import Home from '../pages/Home';
import Domains from '../pages/Domains';
import Profile from '../pages/Profile';
import QueryLogs from '../pages/QueryLogs';
import Setup from '../pages/Setup';
import SetupUserCreation from '../pages/SetupUserCreation';
import SetupPiholes from '../pages/SetupPiholes';
import Login from '../pages/Login';

export const router = createBrowserRouter([
	{
		Component: ProtectedRouteFullInit,
		children: [
			{path: '/', Component: Home},
			{path: '/domains', Component: Domains},
			{path: '/profile', Component: Profile},
			{path: '/query', Component: QueryLogs},
		]
	},
	{
		path: '/setup',
		Component: ProtectedRoute,
		children: [
			{index: true, Component: Setup},
			{path: 'piholes', Component: SetupPiholes}
		]
	},
	{
		Component: UnprotectedRoute,
		children: [
			{path: '/login', Component: Login}
		]
	},
	{
		Component: UnprotectedRouteUninitialized,
		children: [
			{path: '/setup/user', Component: SetupUserCreation}
		]
	},
]);
