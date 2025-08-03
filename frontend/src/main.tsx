import { StrictMode } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from "react-router";
import { router } from './app/router';
import './styles/globals.scss';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>
);
