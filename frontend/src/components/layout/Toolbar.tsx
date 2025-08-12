import { ShieldCheck } from 'lucide-react';
import { useMemo } from 'react';
import { useLocation, Link } from 'react-router';
import '../../styles/components/layout/toolbar.scss';

export default function Toolbar() {
	const { pathname } = useLocation();
	const title = useMemo(() => {
		if (pathname.startsWith('/query')) return 'Query Logs';
		if (pathname.startsWith('/domains')) return 'Domains';
		if (pathname.startsWith('/settings')) return 'Settings';
		return 'Home';
	}, [pathname]);
	return (
		<header className='app-toolbar'>
			<div className='left'>
				<Link to='/' className='brand'>
					<ShieldCheck size={18} />
					<span>Pi-hole Cluster Admin</span>
				</Link>
				<div className='page-title'>{title}</div>
			</div>
		</header>
	);
}
