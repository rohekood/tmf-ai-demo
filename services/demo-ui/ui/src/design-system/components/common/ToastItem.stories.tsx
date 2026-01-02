import type { Meta, StoryObj } from '@storybook/react';
import { ToastItem } from './ToastItem';

const meta = {
    title: 'Design System/Common/ToastItem',
    component: ToastItem,
    tags: ['autodocs'],
    args: {
        onClose: () => { },
    },
    argTypes: {
        onClose: { action: 'closed' },
    },
} satisfies Meta<typeof ToastItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Success: Story = {
    args: {
        toast: {
            id: '1',
            message: 'Operation completed successfully',
            type: 'success',
        },
    },
};

export const Error: Story = {
    args: {
        toast: {
            id: '2',
            message: 'An error occurred while processing your request',
            type: 'error',
        },
    },
};

export const Info: Story = {
    args: {
        toast: {
            id: '3',
            message: 'This is an informational message',
            type: 'info',
        },
    },
};
