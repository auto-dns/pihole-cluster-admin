export interface RouteHandler {
	layoutOptions?: LayoutOptions;
}

export interface LayoutOptions {
	showToolbar?: boolean;
	showSidebar?: boolean;
	pageTitle?: string;
}
