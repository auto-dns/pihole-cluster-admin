// ✅ SVGR components using the ?react query
// (do NOT redeclare '*.svg' here; Vite already declares it as a URL string)
declare module '*.svg?react' {
	import * as React from 'react';
	const ReactComponent: React.FC<React.SVGProps<SVGSVGElement>>;
	export default ReactComponent;
}
