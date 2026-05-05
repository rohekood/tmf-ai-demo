import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import OrderStepTracker from './OrderStepTracker';

describe('OrderStepTracker', () => {
    it('renders all three steps', () => {
        render(<OrderStepTracker />);
        expect(screen.getByText('Inventory')).toBeInTheDocument();
        expect(screen.getByText('Payment')).toBeInTheDocument();
        expect(screen.getByText('Order Created')).toBeInTheDocument();
    });

    it('shows all steps as pending when no status provided', () => {
        render(<OrderStepTracker />);
        expect(screen.getByTestId('step-inventory')).toHaveAttribute('data-status', 'pending');
        expect(screen.getByTestId('step-payment')).toHaveAttribute('data-status', 'pending');
        expect(screen.getByTestId('step-order')).toHaveAttribute('data-status', 'pending');
    });

    it('highlights inventory as current when STARTED', () => {
        render(<OrderStepTracker sagaStatus="STARTED" />);
        expect(screen.getByTestId('step-inventory')).toHaveAttribute('data-status', 'current');
        expect(screen.getByTestId('step-payment')).toHaveAttribute('data-status', 'pending');
        expect(screen.getByTestId('step-order')).toHaveAttribute('data-status', 'pending');
        expect(screen.getByText('In progress...')).toBeInTheDocument();
    });

    it('marks inventory completed and highlights payment as current when INVENTORY_RESERVED', () => {
        render(<OrderStepTracker sagaStatus="INVENTORY_RESERVED" />);
        expect(screen.getByTestId('step-inventory')).toHaveAttribute('data-status', 'completed');
        expect(screen.getByTestId('step-payment')).toHaveAttribute('data-status', 'current');
        expect(screen.getByTestId('step-order')).toHaveAttribute('data-status', 'pending');
    });

    it('marks inventory and payment completed and highlights order as current when PAYMENT_AUTHORIZED', () => {
        render(<OrderStepTracker sagaStatus="PAYMENT_AUTHORIZED" />);
        expect(screen.getByTestId('step-inventory')).toHaveAttribute('data-status', 'completed');
        expect(screen.getByTestId('step-payment')).toHaveAttribute('data-status', 'completed');
        expect(screen.getByTestId('step-order')).toHaveAttribute('data-status', 'current');
    });

    it('marks all steps completed when COMPLETED', () => {
        render(<OrderStepTracker sagaStatus="COMPLETED" />);
        expect(screen.getByTestId('step-inventory')).toHaveAttribute('data-status', 'completed');
        expect(screen.getByTestId('step-payment')).toHaveAttribute('data-status', 'completed');
        expect(screen.getByTestId('step-order')).toHaveAttribute('data-status', 'completed');
    });

    it('marks inventory as failed when INVENTORY_FAILED', () => {
        render(<OrderStepTracker sagaStatus="INVENTORY_FAILED" />);
        expect(screen.getByTestId('step-inventory')).toHaveAttribute('data-status', 'failed');
        expect(screen.getByTestId('step-payment')).toHaveAttribute('data-status', 'pending');
        expect(screen.getByTestId('step-order')).toHaveAttribute('data-status', 'pending');
    });

    it('marks payment as failed when PAYMENT_FAILED', () => {
        render(<OrderStepTracker sagaStatus="PAYMENT_FAILED" />);
        expect(screen.getByTestId('step-inventory')).toHaveAttribute('data-status', 'completed');
        expect(screen.getByTestId('step-payment')).toHaveAttribute('data-status', 'failed');
        expect(screen.getByTestId('step-order')).toHaveAttribute('data-status', 'pending');
    });

    it('marks order step as failed when FAILED', () => {
        render(<OrderStepTracker sagaStatus="FAILED" />);
        expect(screen.getByTestId('step-order')).toHaveAttribute('data-status', 'failed');
    });
});
