import { ReactComponent as LogoMark } from '@/assets/brand/logo-primary.svg';

export function Logo({ color = 'currentColor', size = 24 }: { color?: string; size?: number }) {
	return (
		<LogoMark width={size} height={size} style={{ color }} aria-label='Pi-hole Cluster Admin' />
	);
}
