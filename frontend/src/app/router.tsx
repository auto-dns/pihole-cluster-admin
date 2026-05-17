import { createBrowserRouter } from 'react-router';
// Routes
import { ProtectedRouteFullInit, ProtectedRouteUninitialized } from './routes/ProtectedRoute';
import { UnprotectedRoute, UnprotectedRouteUninitialized } from './routes/UnprotectedRoute';
// Pages
import { AppLayout } from '@/components/Layout/AppLayout';
import { Home } from '@/pages/Home';
import { Blocking } from '@/pages/Blocking/Blocking';
import { Domains } from '@/pages/Domains';
import { QueryLogs } from '@/pages/QueryLogs';
import { RecentBlocks } from '@/pages/RecentBlocks';
import Setup from '@/pages/Setup/Setup';
import { SetupUserCreation } from '@/pages/Setup/SetupUserCreation';
import { SetupPiholes } from '@/pages/Setup/SetupPiholes';
import { Login } from '@/pages/Login';
import { UnhandledRoute } from './routes/UnhandledRoute';
import { Account } from '@/pages/Account/Account';
import { Audit } from '@/pages/Audit';
import { Stats } from '@/pages/Stats';

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
						path: '/blocking',
						Component: Blocking,
						handle: { layoutOptions: { pageTitle: 'Blocking' } },
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
						path: '/recent-blocks',
						Component: RecentBlocks,
						handle: { layoutOptions: { pageTitle: 'Recent Blocks' } },
					},
					{
						path: '/stats',
						Component: Stats,
						handle: { layoutOptions: { pageTitle: 'Stats' } },
					},
					{
						path: '/audit',
						Component: Audit,
						handle: { layoutOptions: { pageTitle: 'Audit Log' } },
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
