import { StrictMode } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router';
import { router } from './app/router';
import { AuthProvider } from './providers/AuthProvider';
import { InitStatusProvider } from './providers/InitializationStatusProvider';
import './styles/globals.scss';
import { PiholeProvider } from './providers/PiholeProvider';

ReactDOM.createRoot(document.getElementById('root')!).render(
	<StrictMode>
		<AuthProvider>
			<InitStatusProvider>
				<PiholeProvider>
					<RouterProvider router={router} />
				</PiholeProvider>
			</InitStatusProvider>
		</AuthProvider>
	</StrictMode>,
);
