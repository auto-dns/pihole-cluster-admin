import { createBrowserRouter } from 'react-router';
import Login from '../pages/Login';

export const router = createBrowserRouter([
	{
		path: "/",
		element: <div>Hello World</div>,
	},
	{
		path: "/login",
		element: <Login/>,
	},
	{
		path: "/setup",
		element: <div>Setup</div>,
	},
	{
		path: "/query",
		element: <div>Query</div>,
	},
	{
		path: "/profile",
		element: <div>Profile</div>,
	},
	{
		path: "/domains",
		element: <div>Domains</div>,
	},
]);
