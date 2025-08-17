import { ShieldCheck, Menu } from 'lucide-react';
import { useMemo } from 'react';
import { useLocation, Link } from 'react-router';
import { useLayout } from '../../providers/LayoutProvider';
import styles from './Toolbar.module.scss';

export default function Toolbar() {
	const { setSidebarOpen } = useLayout();
	const { pathname } = useLocation();
	const title = useMemo(() => {
		if (pathname.startsWith('/query')) return 'Query Logs';
		if (pathname.startsWith('/domains')) return 'Domains';
		if (pathname.startsWith('/settings')) return 'Settings';
		return 'Home';
	}, [pathname]);

	return (
		<header className={styles.toolbar}>
			<div className={styles.left}>
				{/* mobile-only hamburger menu button */}
				<button
					className={styles.hamburger}
					aria-label='Open navigation'
					onClick={() => setSidebarOpen((v) => !v)}
				>
					<Menu size={16} />
				</button>

				<Link to='/' className={styles.brand}>
					<ShieldCheck size={18} />
					<span>Pi-hole Cluster Admin</span>
				</Link>

				<div className={styles.title}>{title}</div>
			</div>
		</header>
	);
}
