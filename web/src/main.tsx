import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { router } from './router.tsx';
import './index.css';

const el = document.getElementById('root');
if (!el) {
  throw new Error('root element missing');
}

createRoot(el).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
);
