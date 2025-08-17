import { Outlet, UIMatch, useMatches } from 'react-router';
import { RouteHandler } from '../../types/layout';
import Toolbar from './Toolbar';
import Sidebar from './Sidebar';
import styles from './AppLayout.module.scss';
import { LayoutProvider } from '@/providers/LayoutProvider';

const DEFAULT_LAYOUT_OPTIONS = {
	showToolbar: true,
	showSidebar: true,
};

export default function AppLayout() {
	const matches = useMatches() as UIMatch<unknown, RouteHandler>[];
	const layoutOptions = matches.reduce((acc, match) => {
		const handle = match.handle as RouteHandler | undefined;
		return handle?.layoutOptions ? { ...acc, ...handle.layoutOptions } : acc;
	}, DEFAULT_LAYOUT_OPTIONS);

	return (
		<LayoutProvider>
			<div className={styles.layout}>
				{layoutOptions?.showToolbar && <Toolbar />}
				<div className={styles.main}>
					{layoutOptions?.showSidebar && <Sidebar />}
					<div className={styles.content}>
						<Outlet />
					</div>
				</div>
			</div>
		</LayoutProvider>
	);
}
