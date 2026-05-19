import { NavLink } from 'react-router';
import {
	BarChart2,
	ChevronRight,
	ChevronLeft,
	Clock,
	Database,
	FileText,
	Home,
	List,
	Monitor,
	Search,
	Shield,
	ShieldAlert,
	SettingsIcon,
	X,
	User,
	LogOut,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import classNames from 'classnames';
import { useLayout } from '@/providers/LayoutProvider';
import { useAuth } from '@/providers/AuthProvider';
import { useClusterOverview } from '@/hooks/useClusterOverview';
import { ClusterHeader } from '@/components/ClusterHeader';
import styles from './Sidebar.module.scss';

type NavLinkItem = {
	to: string;
	label: string;
	icon: LucideIcon;
	end?: boolean;
};

type NavSection = {
	label: string;
	links: NavLinkItem[];
};

const homeLink: NavLinkItem = { to: '/', label: 'Home', icon: Home, end: true };

const sections: NavSection[] = [
	{
		label: 'Monitoring',
		links: [
			{ to: '/stats', label: 'Stats', icon: BarChart2 },
			{ to: '/query', label: 'Query Log', icon: FileText },
			{ to: '/recent-blocks', label: 'Recent Blocks', icon: ShieldAlert },
			{ to: '/diagnose', label: 'Site Diagnoser', icon: Search },
		],
	},
	{
		label: 'Management',
		links: [
			{ to: '/domains', label: 'Domains', icon: List },
			{ to: '/adlists', label: 'Adlists', icon: Database },
			{ to: '/clients', label: 'Clients', icon: Monitor },
		],
	},
	{
		label: 'Cluster',
		links: [
			{ to: '/blocking', label: 'Blocking', icon: Shield },
			{ to: '/audit', label: 'Audit Log', icon: Clock },
		],
	},
	{
		label: 'System',
		links: [
			{ to: '/settings', label: 'Settings', icon: SettingsIcon },
		],
	},
];

const accountLinks: NavLinkItem[] = [{ to: '/account', label: 'Account', icon: User }];

export function Sidebar() {
	const { logout } = useAuth();
	const { isMobile, sidebarOpen: open, setSidebarOpen: setOpen } = useLayout();
	const clusterOverview = useClusterOverview();

	function renderLink({ to, label, icon: Icon, end }: NavLinkItem) {
		return (
			<NavLink
				key={to}
				to={to}
				end={end}
				className={({ isActive }) =>
					classNames(styles.navItem, styles.noUnderline, {
						[styles.active]: isActive,
					})
				}
				title={!open ? label : undefined}
				aria-label={!open ? label : undefined}
				onClick={() => {
					if (isMobile) setOpen(false);
				}}
			>
				<Icon size={18} className={styles.icon} />
				<span className={styles.label}>{label}</span>
			</NavLink>
		);
	}

	return (
		<>
			<aside
				id='sidebar'
				className={classNames(styles.sidebar, { [styles.collapsed]: !open })}
				aria-label='Primary navigation'
			>
				<div className={styles.header}>
					<div className={styles.headerGrid}>
						{open && (
							<>
								<div aria-hidden />
								<NavLink
									key='brand-link'
									to='/'
									className={classNames(
										styles.brandTitle,
										styles.navItem,
										styles.noUnderline,
									)}
									title={!open ? 'Pi-hole Cluster Admin' : undefined}
									aria-label={!open ? 'Pi-hole Cluster Admin' : undefined}
									onClick={() => {
										if (isMobile) setOpen(false);
									}}
								>
									Pi-hole Cluster
									<br />
									Admin
								</NavLink>
							</>
						)}
						<button
							className={classNames(styles.toggleButton, { [styles.closed]: !open })}
							onClick={() => setOpen((v) => !v)}
							aria-label='Collapse sidebar'
							title='Collapse'
						>
							{open ? (
								isMobile ? (
									<X size={16} />
								) : (
									<ChevronLeft size={16} />
								)
							) : (
								<ChevronRight size={16} />
							)}
						</button>
					</div>

					<ClusterHeader open={open} clusterOverview={clusterOverview} />
				</div>

				<nav className={styles.nav}>
					{renderLink(homeLink)}

					<div className={styles.navDivider} />

					{sections.map((section) => (
						<div key={section.label} className={styles.navSection}>
							<span className={styles.navSectionLabel}>{section.label}</span>
							{section.links.map(renderLink)}
						</div>
					))}
				</nav>

				<div className={styles.bottom}>
					{accountLinks.map(({ to, label, icon: Icon }) => (
						<NavLink
							key={to}
							to={to}
							className={({ isActive }) =>
								classNames(styles.navItem, styles.noUnderline, styles.accountItem, {
									[styles.active]: isActive,
								})
							}
							title={!open ? label : undefined}
							aria-label={!open ? label : undefined}
							onClick={() => {
								if (isMobile) setOpen(false);
							}}
						>
							<Icon size={18} className={styles.icon} />
							<span className={styles.label}>{label}</span>
						</NavLink>
					))}

					<button
						type='button'
						className={classNames(styles.navItem, styles.logoutBtn)}
						onClick={logout}
						title={!open ? 'Logout' : undefined}
						aria-label={!open ? 'Logout' : undefined}
					>
						<LogOut size={18} className={styles.icon} />
						<span className={styles.label}>Logout</span>
					</button>
				</div>
			</aside>

			{/* Mobile backdrop */}
			{isMobile && (
				<div
					className={classNames(styles.backdrop, { [styles.show]: open })}
					onClick={() => setOpen(false)}
					aria-hidden='true'
				/>
			)}
		</>
	);
}
