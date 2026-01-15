import type { Meta, StoryObj } from '@storybook/react';
import CategoryPicker from './CategoryPicker';
import type { Category } from '../types';

const meta: Meta<typeof CategoryPicker> = {
    title: 'Catalog/Components/CategoryPicker',
    component: CategoryPicker,
    parameters: {
        layout: 'centered',
    },
    tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof CategoryPicker>;

const mockCategories: Category[] = [
    {
        id: '1',
        name: 'Smartphones',
        description: 'Mobile phones',
        isRoot: true,
        validFor: {},
        lastUpdate: '2023-01-01',
        lifecycleStatus: 'Active',
    },
    {
        id: '2',
        name: 'Laptops',
        description: 'Portable computers',
        isRoot: true,
        validFor: {},
        lastUpdate: '2023-01-01',
        lifecycleStatus: 'Active',
    },
    {
        id: '3',
        name: 'Accessories',
        description: 'Various accessories',
        isRoot: false,
        parentId: '1',
        validFor: {},
        lastUpdate: '2023-01-01',
        lifecycleStatus: 'Active',
    },
    {
        id: '4',
        name: 'Tablets (Retired)',
        description: 'Old tablets',
        isRoot: true,
        validFor: {},
        lastUpdate: '2022-01-01',
        lifecycleStatus: 'Retired',
    },
    {
        id: '5',
        name: 'Other Catalog Cat',
        description: 'Belongs to another catalog',
        isRoot: true,
        catalogId: 'other-cat-id',
        validFor: {},
        lastUpdate: '2023-01-01',
        lifecycleStatus: 'Active',
    }
];

export const DefaultTags: Story = {
    args: {
        selectedIds: ['1'],
        categories: mockCategories,
        variant: 'tags',
        onChange: (ids) => console.log('onChange', ids),
    },
};

export const ListView: Story = {
    args: {
        selectedIds: ['1', '4'],
        categories: mockCategories,
        variant: 'list',
        onChange: (ids) => console.log('onChange', ids),
    },
};

export const EmptyList: Story = {
    args: {
        selectedIds: [],
        categories: mockCategories,
        variant: 'list',
        onChange: (ids) => console.log('onChange', ids),
    },
};

export const WithCatalogFilter: Story = {
    args: {
        selectedIds: [],
        categories: mockCategories,
        catalogId: 'other-cat-id',
        variant: 'tags',
        onChange: (ids) => console.log('onChange', ids),
    },
    render: (args) => (
        <div>
            <p className="text-muted mb-2">Filtering to only show categories belonging to 'other-cat-id' (should see 'Other Catalog Cat')</p>
            <CategoryPicker {...args} />
        </div>
    )
};
