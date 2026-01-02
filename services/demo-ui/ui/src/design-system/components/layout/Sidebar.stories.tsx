import type { Meta, StoryObj } from '@storybook/react';
import { Sidebar } from './Sidebar';
import { Users, Building2, Bug } from 'lucide-react';

const meta = {
    title: 'Design System/Layout/Sidebar',
    component: Sidebar,
    parameters: {
        layout: 'fullscreen',
    },
    tags: ['autodocs'],
    argTypes: {
        onToggleCollapse: { action: 'toggleCollapse' },
        onLogin: { action: 'login' },
        onLogout: { action: 'logout' },
    },
} satisfies Meta<typeof Sidebar>;

export default meta;
type Story = StoryObj<typeof meta>;

// Mock Render Link
const MockLink = (item: any, children: React.ReactNode, className: string) => (
    <a href={item.path} className={className} onClick={(e) => e.preventDefault()}>
        {children}
    </a>
);

const defaultNavGroups = [
    {
        title: 'Management',
        items: [
            { path: '/parties', label: 'Parties', icon: <Users size={20} />, isActive: true },
            { path: '/customers', label: 'Customers', icon: <Building2 size={20} /> },
        ],
    },
    {
        title: 'Developer',
        items: [
            { path: '/debug', label: 'Debug Console', icon: <Bug size={20} /> },
        ],
    },
];

export const Expanded: Story = {
    args: {
        collapsed: false,
        mobileOpen: false,
        navGroups: defaultNavGroups,
        renderLink: MockLink,
        isAuthenticated: true,
        user: {
            name: 'John Doe',
            email: 'john.doe@example.com',
            picture: 'https://i.pravatar.cc/150?u=a042581f4e29026024d',
        },
        onToggleCollapse: () => { },
    },
};

export const Collapsed: Story = {
    args: {
        ...Expanded.args,
        collapsed: true,
    },
};

export const Loading: Story = {
    args: {
        ...Expanded.args,
        isLoading: true,
    },
};

export const Guest: Story = {
    args: {
        ...Expanded.args,
        isAuthenticated: false,
        user: undefined,
    },
};
