import type { Meta, StoryObj } from '@storybook/react';
import { Footer } from './Footer';

const meta = {
    title: 'Design System/Layout/Footer',
    component: Footer,
    parameters: {
        layout: 'fullscreen',
    },
    tags: ['autodocs'],
} satisfies Meta<typeof Footer>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    args: {
        version: '1.0.0',
        companyName: 'ACME Corp',
    },
};

export const WithLinks: Story = {
    args: {
        version: '2.5.0-beta',
        links: [
            { label: 'Documentation', href: '#' },
            { label: 'Support', href: '#' },
            { label: 'Terms of Service', href: '#' },
        ],
    },
};
