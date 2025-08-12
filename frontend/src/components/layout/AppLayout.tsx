import { Outlet, UIMatch, useMatches } from 'react-router';
import '../../styles/components/layout/app-layout.scss';
import { RouteHandler } from '../../types/layout';
import Toolbar from './Toolbar';
import Sidebar from './Sidebar';

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
		<div className='app-layout'>
			{layoutOptions?.showToolbar && <Toolbar />}
			<div className='app-main'>
				{layoutOptions?.showSidebar && <Sidebar />}
				<div className='app-content'>
					<Outlet />
				</div>
			</div>
		</div>
	);
}
