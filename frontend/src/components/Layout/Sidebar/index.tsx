import { NavLink } from 'react-router';
import {
	ChevronRight,
	ChevronLeft,
	FileText,
	Home,
	List,
	SettingsIcon,
	X,
	User,
	LogOut,
} from 'lucide-react';
import { useLayout } from '../../../providers/LayoutProvider';
import { useAuth } from '../../../providers/AuthProvider';
import { useClusterHealth } from '../../../hooks/useClusterHealth';
import StatusLight from '../../StatusLight/StatusLight';
import Logo from '../../Logo/Logo';
import classNames from 'classnames';
import styles from './index.module.scss';

const links = [
	{ to: '/', label: 'Home', icon: Home, end: true },
	{ to: '/query', label: 'Query Log', icon: FileText },
	{ to: '/domains', label: 'Domains', icon: List },
	{ to: '/settings', label: 'Settings', icon: SettingsIcon },
];

const accountLinks = [{ to: '/account', label: 'Account', icon: User }];

export default function Sidebar() {
	const { logout } = useAuth();
	const { isMobile, sidebarOpen: open, setSidebarOpen: setOpen } = useLayout();
	const { summary } = useClusterHealth();

	const online = summary?.online ?? 0;
	const total = summary?.total ?? 0;
	const statusColor =
		total === 0
			? 'var(--border-primary)'
			: online === 0
				? 'var(--accent-danger)'
				: online !== total
					? 'var(--accent-warn)'
					: 'var(--accent-success)';
	const pulse = total !== 0 && (online === total || online !== 0);

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

					<div className={styles.logoWrap} aria-hidden>
						<Logo className={styles.logo} />
					</div>

					<div
						className={styles.headerStatus}
						data-count={`${online}/${total}`}
						title={`${online}/${total} nodes online`}
					>
						<StatusLight
							label={`${online} of ${total} nodes online`}
							title={`${online} of ${total} nodes online`}
							color={statusColor}
							pulse={pulse}
							durationMs={pulse ? 1800 : 0}
							mode='blink'
						/>
						{open && (
							<>
								<strong>
									{online}/{total}
								</strong>
								<span className={styles.muted}>nodes</span>
							</>
						)}
					</div>
				</div>

				<nav className={styles.nav}>
					{links.map(({ to, label, icon: Icon, end }) => (
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
