import type { Meta, StoryObj } from '@storybook/react';
import { IconButtonArea } from './IconButtonArea';
import { IconButton } from './IconButton';
import { Eye, Edit, Trash2, MoreHorizontal } from 'lucide-react';

const meta = {
    title: 'Design System/Common/IconButtonArea',
    component: IconButtonArea,
    parameters: {
        layout: 'padded',
    },
    tags: ['autodocs'],
    argTypes: {
        alignment: {
            control: 'select',
            options: ['start', 'center', 'end', 'between'],
        },
    },
} satisfies Meta<typeof IconButtonArea>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    args: {
        children: (
            <>
                <IconButton icon={<Eye size={20} />} title="View" />
                <IconButton icon={<Edit size={20} />} title="Edit" />
                <IconButton variant="danger" icon={<Trash2 size={20} />} title="Delete" />
            </>
        ),
    },
};

export const AlignStart: Story = {
    args: {
        alignment: 'start',
        children: (
            <>
                <IconButton variant="ghost" icon={<MoreHorizontal size={20} />} />
                <IconButton icon={<Eye size={20} />} />
            </>
        ),
    },
};

export const WithManyButtons: Story = {
    args: {
        children: (
            <>
                <IconButton icon={<Eye size={20} />} title="View" />
                <IconButton icon={<Edit size={20} />} title="Edit" />
                <IconButton variant="ghost" icon={<MoreHorizontal size={20} />} title="More" />
                <div style={{ width: '1px', height: '24px', background: '#e2e8f0', margin: '0 4px' }} />
                <IconButton variant="danger" icon={<Trash2 size={20} />} title="Delete" />
            </>
        ),
    },
};
