import { NavLink } from 'react-router';
import { ChevronRight, ChevronLeft, FileText, Home, List, SettingsIcon } from 'lucide-react';
import classNames from 'classnames';
import '../../styles/components/layout/sidebar.scss';
import { useLocalStorageState } from '../../hooks/useLocalStorageState';
import { useClusterHealth } from '../../hooks/useClusterHealth';

const links = [
	{ to: '/', label: 'Home', icon: Home, end: true },
	{ to: '/query', label: 'Query Log', icon: FileText },
	{ to: '/domains', label: 'Domains', icon: List },
	{ to: '/settings', label: 'Settings', icon: SettingsIcon },
];

export default function Sidebar() {
	const [collapsed, setCollapsed] = useLocalStorageState<boolean>(
		'pihole-cluster-admin.sidebarCollapsed',
		false,
		{ syncAcrossTabs: true },
	);
	const { summary } = useClusterHealth();

	return (
		<aside className={classNames('app-sidebar', { collapsed })}>
			<button
				className='collapse'
				onClick={() => setCollapsed((v) => !v)}
				aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
				title={collapsed ? 'Expand' : 'Collapse'}
			>
				{collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
			</button>
			<nav>
				{links.map(({ to, label, icon: Icon, end }) => (
					<NavLink
						key={to}
						to={to}
						end={end}
						className={({ isActive }) => classNames('nav-item', { active: isActive })}
						title={collapsed ? label : undefined}
						aria-label={collapsed ? label : undefined}
					>
						<Icon size={18} className='icon' />
						<span className='label'>{label}</span>
					</NavLink>
				))}
			</nav>
			<div className='foot'>
				<div
					className='cluster-mini'
					data-count={`${summary?.online}/${summary?.total}`}
					title={`${summary?.online}/${summary?.total} nodes online`}
					aria-label={`${summary?.online} of ${summary?.total} nodes online`}
				>
					<span className='dot online' />
					<strong>
						{summary?.online}/{summary?.total}
					</strong>{' '}
					<span className='muted'>nodes</span>
				</div>
			</div>
		</aside>
	);
}
