import type { Meta, StoryObj } from '@storybook/react';
import { IconButton } from './IconButton';
import { Edit, Trash2, Eye, Plus, MoreVertical } from 'lucide-react';

const meta = {
    title: 'Design System/Common/IconButton',
    component: IconButton,
    parameters: {
        layout: 'centered',
    },
    tags: ['autodocs'],
    argTypes: {
        variant: {
            control: 'select',
            options: ['default', 'ghost', 'danger', 'primary'],
        },
        size: {
            control: 'select',
            options: ['sm', 'md', 'lg'],
        },
        disabled: {
            control: 'boolean',
        },
    },
} satisfies Meta<typeof IconButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    args: {
        icon: <Eye size={20} />,
        title: 'View Details',
    },
};

export const Primary: Story = {
    args: {
        variant: 'primary',
        icon: <Plus size={20} />,
        title: 'Add New Item',
    },
};

export const Danger: Story = {
    args: {
        variant: 'danger',
        icon: <Trash2 size={20} />,
        title: 'Delete Item',
    },
};

export const Ghost: Story = {
    args: {
        variant: 'ghost',
        icon: <MoreVertical size={20} />,
        title: 'More Options',
    },
};

export const Small: Story = {
    args: {
        size: 'sm',
        icon: <Edit size={16} />,
        title: 'Edit (Small)',
    },
};

export const Large: Story = {
    args: {
        size: 'lg',
        icon: <Edit size={24} />,
        title: 'Edit (Large)',
    },
};

export const Disabled: Story = {
    args: {
        disabled: true,
        icon: <Trash2 size={20} />,
        title: 'Delete (Disabled)',
        variant: 'danger',
    },
};

export const AsLink: Story = {
    args: {
        to: '/some-path',
        icon: <Eye size={20} />,
        title: 'Go to details',
    },
};
