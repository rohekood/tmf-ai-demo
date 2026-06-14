import type { Meta, StoryObj } from '@storybook/react';
import { DateInput } from './DateInput';

const meta = {
    title: 'Design System/Common/DateInput',
    component: DateInput,
    parameters: {
        layout: 'centered',
        backgrounds: { default: 'dark' },
    },
    tags: ['autodocs'],
    argTypes: {
        label: { control: 'text' },
        disabled: { control: 'boolean' },
        value: { control: 'text' },
    },
    decorators: [
        (Story) => (
            <div style={{ width: 280 }}>
                <Story />
            </div>
        ),
    ],
} satisfies Meta<typeof DateInput>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    args: {
        label: 'Start Date',
    },
};

export const Prefilled: Story = {
    args: {
        label: 'Valid From',
        defaultValue: '2026-06-14',
    },
};

export const WithoutLabel: Story = {
    args: {
        'aria-label': 'Date without label',
    },
};

export const Disabled: Story = {
    args: {
        label: 'End Date',
        defaultValue: '2026-06-14',
        disabled: true,
    },
};
