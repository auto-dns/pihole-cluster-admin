import { createBrowserRouter } from 'react-router';
// Routes
import { ProtectedRouteFullInit, ProtectedRouteUninitialized } from './routes/ProtectedRoute';
import { UnprotectedRoute, UnprotectedRouteUninitialized } from './routes/UnprotectedRoute';
// Pages
import AppLayout from '../components/layout/AppLayout';
import Home from '../pages/Home';
import Domains from '../pages/Domains';
import Profile from '../pages/Profile';
import QueryLogs from '../pages/QueryLogs';
import Setup from '../pages/Setup';
import SetupUserCreation from '../pages/SetupUserCreation';
import SetupPiholes from '../pages/SetupPiholes';
import Login from '../pages/Login';
import UnhandledRoute from './routes/UnhandledRoute';

export const router = createBrowserRouter([
	{
		Component: AppLayout,
		children: [
			{
				Component: ProtectedRouteFullInit,
				children: [
					{ path: '/', Component: Home },
					{ path: '/domains', Component: Domains },
					{ path: '/settings', Component: Profile },
					{ path: '/query', Component: QueryLogs },
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
