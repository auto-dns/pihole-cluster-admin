import { ShieldCheck } from 'lucide-react';
import { useMemo } from 'react';
import { useLocation, Link } from 'react-router';
import styles from './Toolbar.module.scss';

export default function Toolbar() {
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
				<Link to='/' className={styles.brand}>
					<ShieldCheck size={18} />
					<span>Pi-hole Cluster Admin</span>
				</Link>
				<div className={styles.title}>{title}</div>
			</div>
		</header>
	);
}
