import { createBrowserRouter } from "react-router";

export const router = createBrowserRouter([
	{
		path: "/",
		element: <div>Hello World</div>,
	},
	{
		path: "/login",
		element: <div>Login</div>,
	},
	{
		path: "/setup",
		element: <div>Setup</div>,
	},
	{
		path: "/query",
		element: <div>Login</div>,
	},
	{
		path: "/profile",
		element: <div>Login</div>,
	},
	{
		path: "/domains",
		element: <div>Domains</div>,
	},
]);
