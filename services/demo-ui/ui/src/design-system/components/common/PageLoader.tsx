import { type ReactNode } from 'react';
import './PageLoader.css';

export function PageLoader(): ReactNode {
    return (
        <div className="page-loader">
            <div className="loader-spinner"></div>
            <p>Loading...</p>
        </div>
    );
}
