import { StrictMode } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router';
import { router } from './app/router';
import { AuthProvider } from './app/AuthProvider';
import { InitStatusProvider } from './app/InitializationStatusProvider';
import './styles/globals.scss';
import './styles/layout/app-layout.scss';

ReactDOM.createRoot(document.getElementById('root')!).render(
	<StrictMode>
		<AuthProvider>
			<InitStatusProvider>
				<RouterProvider router={router} />
			</InitStatusProvider>
		</AuthProvider>
	</StrictMode>,
);
