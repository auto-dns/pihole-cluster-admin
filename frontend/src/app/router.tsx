import { createBrowserRouter } from 'react-router';
// Routes
import { ProtectedRouteFullInit, ProtectedRouteUninitialized } from './routes/ProtectedRoute';
import { UnprotectedRoute, UnprotectedRouteUninitialized } from './routes/UnprotectedRoute';
// Pages
import AppLayout from '../components/Layout/AppLayout';
import Home from '../pages/Home/Home';
import Domains from '../pages/Domains/Domains';
import QueryLogs from '../pages/QueryLogs/QueryLogs';
import Setup from '../pages/Setup/Setup';
import SetupUserCreation from '../pages/Setup/SetupUserCreation/SetupUserCreation';
import SetupPiholes from '../pages/Setup/SetupPiholes/SetupPiholes';
import Login from '../pages/Login/Login';
import UnhandledRoute from './routes/UnhandledRoute';
import Account from '../pages/Account/Account';

export const router = createBrowserRouter([
	{
		Component: AppLayout,
		children: [
			{
				Component: ProtectedRouteFullInit,
				children: [
					{
						path: '/',
						Component: Home,
						handle: { layoutOptions: { pageTitle: 'Home' } },
					},
					{
						path: '/domains',
						Component: Domains,
						handle: { layoutOptions: { pageTitle: 'Domains' } },
					},
					{
						path: '/query',
						Component: QueryLogs,
						handle: { layoutOptions: { pageTitle: 'Query Logs' } },
					},
					{
						path: '/account',
						Component: Account,
						handle: { layoutOptions: { pageTitle: 'Account' } },
					},
				],
			},
			{
				children: [
					{
						Component: ProtectedRouteUninitialized,
						children: [
							{
								path: '/setup',
								Component: Setup,
								children: [{ path: 'piholes', Component: SetupPiholes }],
							},
						],
					},
					{
						Component: UnprotectedRouteUninitialized,
						children: [{ path: '/setup/user', Component: SetupUserCreation }],
					},
					{
						Component: UnprotectedRoute,
						children: [{ path: '/login', Component: Login }],
					},
					{
						path: '*',
						Component: UnhandledRoute,
					},
				],
				handle: { layoutOptions: { showToolbar: false, showSidebar: false } },
			},
		],
	},
]);
