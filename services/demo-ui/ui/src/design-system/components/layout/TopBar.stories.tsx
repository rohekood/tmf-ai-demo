import type { Meta, StoryObj } from '@storybook/react';
import { TopBar } from './TopBar';
import { Bell, Search } from 'lucide-react';

const meta = {
    title: 'Design System/Layout/TopBar',
    component: TopBar,
    parameters: {
        layout: 'fullscreen',
    },
    tags: ['autodocs'],
    argTypes: {
        onToggleMobileMenu: { action: 'toggleMobileMenu' },
    },
} satisfies Meta<typeof TopBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    args: {
        title: 'TMF Demo Dashboard',
        subtitle: 'Managed via Golang BFF & RabbitMQ',
    },
};

export const WithActions: Story = {
    args: {
        title: 'Customer Management',
        subtitle: 'View and manage customer details',
        actions: (
            <>
                <button className="p-2 text-slate-400 hover:text-slate-600 rounded-full hover:bg-slate-100">
                    <Search size={20} />
                </button>
                <button className="p-2 text-slate-400 hover:text-slate-600 rounded-full hover:bg-slate-100">
                    <Bell size={20} />
                </button>
                <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-500 to-purple-500 ml-2"></div>
            </>
        ),
    },
};

export const MobileView: Story = {
    parameters: {
        viewport: {
            defaultViewport: 'mobile1',
        },
    },
    args: {
        title: 'Mobile Dashboard',
        onToggleMobileMenu: () => { },
    },
};
