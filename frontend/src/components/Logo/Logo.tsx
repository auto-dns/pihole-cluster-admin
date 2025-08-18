// Logo.tsx
import LogoMark from '@/assets/brand/logo-primary.svg?react';

type Props = {
	/** Preferred: style via CSS classes; enables smooth transitions */
	className?: string;
	/** Inline style override if needed */
	style?: React.CSSProperties;
	/** Color uses currentColor in the SVG for themeability */
	color?: string;
	/** Back-compat: if provided, sets both width/height (px) */
	size?: number;
	/** Optional explicit width/height overrides */
	width?: number | string;
	height?: number | string;
	/** Accessible label (falls back to name) */
	title?: string;
};

export function Logo({
	className,
	style,
	color = 'currentColor',
	size,
	width,
	height,
	title = 'Pi-hole Cluster Admin',
}: Props) {
	// Build props: prefer explicit width/height if provided; else `size`; else CSS controls it.
	const dimProps: Record<string, unknown> = {};
	if (width != null) dimProps.width = width;
	if (height != null) dimProps.height = height;
	if (size != null && width == null && height == null) {
		dimProps.width = size;
		dimProps.height = size;
	}

	return (
		<LogoMark
			{...dimProps}
			className={className}
			style={{ color, ...style }}
			aria-label={title}
			role='img'
			focusable='false'
		/>
	);
}
