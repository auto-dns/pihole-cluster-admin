declare module '*.svg' {
	import * as React from 'react';
	export const ReactComponent: React.FC<React.SVGProps<SVGSVGElement>>;
	const src: string; // default export is still the URL string
	export default src;
}
