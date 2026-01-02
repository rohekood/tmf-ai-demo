import type { Meta, StoryObj } from '@storybook/react';
import { ButtonArea } from './ButtonArea';

const meta = {
    title: 'Design System/Common/ButtonArea',
    component: ButtonArea,
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
} satisfies Meta<typeof ButtonArea>;

export default meta;
type Story = StoryObj<typeof meta>;

const ButtonPlaceholder = ({ children, primary = false }: { children: React.ReactNode; primary?: boolean }) => (
    <button
        className={`px-4 py-2 rounded-md font-medium transition-colors ${primary
                ? 'bg-blue-600 text-white hover:bg-blue-700'
                : 'bg-white text-slate-700 border border-slate-300 hover:bg-slate-50'
            }`}
    >
        {children}
    </button>
);

export const Default: Story = {
    args: {
        children: (
            <>
                <ButtonPlaceholder>Cancel</ButtonPlaceholder>
                <ButtonPlaceholder primary>Save Changes</ButtonPlaceholder>
            </>
        ),
    },
};

export const ALignStart: Story = {
    args: {
        alignment: 'start',
        children: (
            <>
                <ButtonPlaceholder primary>Add Item</ButtonPlaceholder>
            </>
        ),
    },
};

export const AlignCenter: Story = {
    args: {
        alignment: 'center',
        children: (
            <>
                <ButtonPlaceholder>Previous</ButtonPlaceholder>
                <ButtonPlaceholder primary>Next</ButtonPlaceholder>
            </>
        ),
    },
};

export const ManyButtons: Story = {
    args: {
        children: (
            <>
                <ButtonPlaceholder>Option 1</ButtonPlaceholder>
                <ButtonPlaceholder>Option 2</ButtonPlaceholder>
                <ButtonPlaceholder>Option 3</ButtonPlaceholder>
                <ButtonPlaceholder primary>Confirm</ButtonPlaceholder>
            </>
        ),
    },
};
