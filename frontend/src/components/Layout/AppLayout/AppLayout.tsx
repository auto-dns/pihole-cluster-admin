import { Outlet, UIMatch, useMatches } from 'react-router';
import { RouteHandler } from '@/types/layout';
import { LayoutProvider, useLayout } from '@/providers/LayoutProvider';
import { ClusterOverviewProvider } from '@/providers/ClusterOverviewProvider';
import { Toolbar } from '@/components/Layout/Toolbar';
import { Sidebar } from '@/components/Layout/Sidebar';
import styles from './AppLayout.module.scss';

const DEFAULT_LAYOUT_OPTIONS = {
	showToolbar: true,
	showSidebar: true,
	pageTitle: undefined as string | undefined, // optional override support
};

export function AppLayout() {
	const matches = useMatches() as UIMatch<unknown, RouteHandler>[];
	const layoutOptions = matches.reduce((acc, match) => {
		const handle = match.handle as RouteHandler | undefined;
		return handle?.layoutOptions ? { ...acc, ...handle.layoutOptions } : acc;
	}, DEFAULT_LAYOUT_OPTIONS);

	const { isMobile, sidebarOpen } = useLayout();

	return (
		<LayoutProvider>
			<ClusterOverviewProvider>
				<div
					className={styles.layout}
					data-collapsed={!isMobile && !sidebarOpen}
					data-open={isMobile && sidebarOpen}
				>
					{layoutOptions.showSidebar && <Sidebar />}

					<div className={styles.rightPane}>
						{layoutOptions.showToolbar && <Toolbar pageTitle={layoutOptions.pageTitle} />}
						<main className={styles.content} role='main' id='main'>
							<Outlet />
						</main>
					</div>
				</div>
			</ClusterOverviewProvider>
		</LayoutProvider>
	);
}
