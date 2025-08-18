import { Menu } from 'lucide-react';
import { useMemo } from 'react';
import { useLocation } from 'react-router';
import { useLayout } from '../../../providers/LayoutProvider';
import styles from './index.module.scss';

type Props = {
	pageTitle?: string;
};

export default function Toolbar({ pageTitle }: Props) {
	const { sidebarOpen, setSidebarOpen } = useLayout();
	const { pathname } = useLocation();

	const fallbackTitle = useMemo(() => {
		if (pathname.startsWith('/query')) return 'Query Logs';
		if (pathname.startsWith('/domains')) return 'Domains';
		if (pathname.startsWith('/settings')) return 'Settings';
		return 'Home';
	}, [pathname]);

	const title = pageTitle ?? fallbackTitle;

	return (
		<header className={styles.toolbar} role='banner'>
			<div className={styles.left}>
				<button
					className={styles.hamburger}
					aria-label={sidebarOpen ? 'Close navigation' : 'Open navigation'}
					aria-controls='sidebar'
					aria-expanded={sidebarOpen}
					onClick={() => setSidebarOpen((v) => !v)}
				>
					<Menu size={16} />
				</button>

				<div className={styles.title}>{title}</div>
			</div>
		</header>
	);
}
