import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import './Layout.css';

export function Layout() {
    return (
        <div className="layout">
            <Sidebar />
            <div className="layout-main">
                <header className="layout-header">
                    <h1>TMF Demo Dashboard</h1>
                    <p className="layout-subtitle">Managed via Golang BFF & RabbitMQ</p>
                </header>
                <main className="layout-content">
                    <Outlet />
                </main>
            </div>
        </div>
    );
}
