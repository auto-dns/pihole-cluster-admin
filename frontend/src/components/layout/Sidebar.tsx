import { NavLink } from 'react-router';
import { FileText, Home, List, SettingsIcon } from 'lucide-react';
import { useState } from 'react';
import classNames from 'classnames';
import '../../styles/components/layout/sidebar.scss';

const links = [
	{ to: '/', label: 'Home', icon: Home, end: true },
	{ to: '/query', label: 'Query Log', icon: FileText },
	{ to: '/domains', label: 'Domains', icon: List },
	{ to: '/settings', label: 'Settings', icon: SettingsIcon },
];

export default function Sidebar() {
	const [collapsed, setCollapsed] = useState<boolean>(false);

	return (
		<aside className={classNames('app-sidebar', { collapsed })}>
			<button
				className='collapse'
				onClick={() => setCollapsed((v) => !v)}
				aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
			>
				{collapsed ? '›' : '‹'}
			</button>
			<nav>
				{links.map(({ to, label, icon: Icon, end }) => (
					<NavLink
						key={to}
						to={to}
						end={end}
						className={({ isActive }) => classNames('nav-item', { active: isActive })}
					>
						<Icon size={18} className='icon' />
						<span className='label'>{label}</span>
					</NavLink>
				))}
			</nav>
			<div className='foot'>
				<div className='cluster-mini'>
					<span className='dot online' />
					{/* TODO: replace this with something dynamic */}
					<span>3/3 nodes</span>
				</div>
			</div>
		</aside>
	);
}
