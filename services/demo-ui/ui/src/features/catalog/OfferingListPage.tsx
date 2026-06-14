import { useState, useMemo, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
    createColumnHelper,
    getCoreRowModel,
    getSortedRowModel,
    flexRender,
    type SortingState,
} from '@tanstack/react-table';
import { useReactTable } from '../../hooks/useReactTable';
import { Plus, Search, ChevronUp, ChevronDown, Eye, Edit, Trash2, Loader2, ShoppingCart } from 'lucide-react';
import { EmptyState } from '../../design-system/components/common/EmptyState';
import { useOfferings, useDeleteOffering } from './api';
import type { ProductOffering } from './types';
import './SpecificationListPage.css'; // Reuse common catalog styles
import { IconButton } from '../../design-system/components/common/IconButton';
import { IconButtonArea } from '../../design-system/components/common/IconButtonArea';

const columnHelper = createColumnHelper<ProductOffering>();

export default function OfferingListPage() {
    const [searchTerm, setSearchTerm] = useState('');
    const [sorting, setSorting] = useState<SortingState>([]);

    const { data: offerings = [], isLoading, error } = useOfferings();
    const deleteMutation = useDeleteOffering();

    const handleDelete = useCallback((id: string, name: string) => {
        if (confirm(`Are you sure you want to delete offering "${name}"?`)) {
            deleteMutation.mutate(id);
        }
    }, [deleteMutation]);

    const filteredOfferings = useMemo(() => {
        if (!searchTerm) return offerings;
        const lowSearch = searchTerm.toLowerCase();
        return offerings.filter(o =>
            o.name.toLowerCase().includes(lowSearch) ||
            o.description?.toLowerCase().includes(lowSearch)
        );
    }, [offerings, searchTerm]);

    const columns = useMemo(
        () => [
            columnHelper.accessor('name', {
                header: 'Name',
                cell: (info) => (
                    <div className="spec-info">
                        <span className="spec-name">{info.getValue()}</span>
                        {info.row.original.isSellable && (
                            <span className="badge badge-success-outline ml-2">Sellable</span>
                        )}
                    </div>
                ),
            }),
            columnHelper.accessor('lifecycleStatus', {
                header: 'Status',
                cell: (info) => (
                    <span className={`status-badge ${info.getValue().toLowerCase()}`}>
                        {info.getValue()}
                    </span>
                ),
                size: 120,
            }),
            columnHelper.accessor((row) => row.productOfferingPrice?.length || 0, {
                id: 'prices',
                header: 'Prices',
                cell: (info) => <span className="count-text">{info.getValue()} prices</span>,
                size: 100,
            }),
            columnHelper.display({
                id: 'actions',
                header: '',
                cell: (info) => (
                    <IconButtonArea alignment="end">
                        <IconButton
                            to={`/catalog/offerings/${info.row.original.id}`}
                            icon={<Eye size={16} />}
                            title="View"
                        />
                        <IconButton
                            to={`/catalog/offerings/${info.row.original.id}/edit`}
                            icon={<Edit size={16} />}
                            title="Edit"
                        />
                        <IconButton
                            variant="danger"
                            title="Delete"
                            onClick={() => handleDelete(info.row.original.id, info.row.original.name)}
                            disabled={deleteMutation.isPending}
                            icon={<Trash2 size={16} />}
                        />
                    </IconButtonArea>
                ),
                size: 120,
            }),
        ],
        [deleteMutation.isPending, handleDelete]
    );

    const table = useReactTable({
        data: filteredOfferings,
        columns,
        state: { sorting },
        onSortingChange: setSorting,
        getCoreRowModel: getCoreRowModel(),
        getSortedRowModel: getSortedRowModel(),
    });

    return (
        <div className="specification-list-page">
            <div className="page-header">
                <div className="page-header-content">
                    <h1 className="page-title">Product Offerings</h1>
                    <p className="page-description">Commercial versions of your products available to customers</p>
                </div>
                <Link to="/catalog/offerings/new" className="btn btn-primary">
                    <Plus size={18} />
                    <span>New Offering</span>
                </Link>
            </div>

            <div className="search-bar card">
                <Search size={20} className="search-icon" />
                <input
                    type="text"
                    placeholder="Search by name or description..."
                    className="search-input"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                />
            </div>

            {error ? (
                <div className="card error-card" role="alert">
                    <p>Failed to load offerings: {error.message}</p>
                </div>
            ) : isLoading ? (
                <div className="card loading-card" role="status">
                    <Loader2 className="spin" size={24} />
                    <p>Loading offerings...</p>
                </div>
            ) : filteredOfferings.length === 0 ? (
                <EmptyState
                    icon={<ShoppingCart size={48} />}
                    title={searchTerm ? 'No offerings match your search.' : 'No product offerings found.'}
                    action={
                        !searchTerm ? (
                            <Link to="/catalog/offerings/new" className="btn btn-primary">
                                Create your first offering
                            </Link>
                        ) : undefined
                    }
                />
            ) : (
                <div className="card table-card">
                    <table className="data-table">
                        <thead>
                            {table.getHeaderGroups().map((headerGroup) => (
                                <tr key={headerGroup.id}>
                                    {headerGroup.headers.map((header) => (
                                        <th
                                            key={header.id}
                                            onClick={header.column.getToggleSortingHandler()}
                                            className={header.column.getCanSort() ? 'sortable' : ''}
                                            style={{ width: header.getSize() }}
                                        >
                                            <div className="th-content">
                                                {flexRender(header.column.columnDef.header, header.getContext())}
                                                {header.column.getIsSorted() === 'asc' && <ChevronUp size={14} />}
                                                {header.column.getIsSorted() === 'desc' && <ChevronDown size={14} />}
                                            </div>
                                        </th>
                                    ))}
                                </tr>
                            ))}
                        </thead>
                        <tbody>
                            {table.getRowModel().rows.map((row) => (
                                <tr key={row.id}>
                                    {row.getVisibleCells().map((cell) => (
                                        <td key={cell.id}>
                                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                        </td>
                                    ))}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                    <div className="table-footer">
                        <span className="table-count">{filteredOfferings.length} offerings</span>
                    </div>
                </div>
            )}
        </div>
    );
}
