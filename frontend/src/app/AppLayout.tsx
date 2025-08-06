import { Outlet, UIMatch, useMatches } from 'react-router';
import '../styles/layout/app-layout.scss';
import { RouteHandler } from '../types/layout';

const DEFAULT_LAYOUT_OPTIONS = {
	showToolbar: true,
	showSidebar: true,
};

export default function AppLayout() {
	const matches = useMatches() as UIMatch<unknown, RouteHandler>[];
	const route = matches[matches.length - 1];
	const layoutOptions = matches.reduce((acc, match) => {
		const handle = match.handle as RouteHandler | undefined;
		return handle?.layoutOptions ? { ...acc, ...handle.layoutOptions } : acc;
	}, DEFAULT_LAYOUT_OPTIONS);

	console.log(matches, route, layoutOptions, route?.handle?.layoutOptions);

	return (
		<div className='app-layout'>
			{layoutOptions?.showToolbar && <div className='app-toolbar'></div>}
			<div className='app-main'>
				{layoutOptions?.showSidebar && <div className='app-sidebar'></div>}
				<div className='app-content'>
					<Outlet />
				</div>
			</div>
		</div>
	);
}
