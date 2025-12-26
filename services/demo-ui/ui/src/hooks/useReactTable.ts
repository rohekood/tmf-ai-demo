import { useReactTable as useReactTableOriginal, type TableOptions, type Table } from '@tanstack/react-table';

// This wrapper isolates the TanStack Table hook from the React Compiler
// to avoid "incompatible-library" warnings, as the library returns mutable
// objects that cannot be safely memoized by the compiler.
export function useReactTable<TData>(options: TableOptions<TData>): Table<TData> {
    "use no memo";
    // eslint-disable-next-line react-hooks/incompatible-library
    return useReactTableOriginal(options);
}
