<script lang="ts">
	import { notify } from '@/lib/util/notify';
	import Spinner from '@/lib/Spinner.svelte';
	import MatrixTroubleshooting from './MatrixTroubleshooting.svelte';
	import axios from 'axios';

	interface ValidationResult {
		check: string;
		status: 'success' | 'warning' | 'error';
		message: string;
		details?: string;
	}

	interface ValidationResponse {
		overallStatus: 'success' | 'warning' | 'error';
		results: ValidationResult[];
		summary: string;
	}

	let { onClose }: { onClose: () => void } = $props();

	let validationLoading = $state(false);
	let validationResults = $state<ValidationResponse | null>(null);
	let showDetails = $state<Record<string, boolean>>({});
	let troubleshootingModalOpen = $state(false);

	async function runValidation() {
		validationLoading = true;
		try {
			const response = await axios.get('/matrix/validate');
			validationResults = response.data;
		} catch (error) {
			console.error('Failed to run Matrix validation:', error);
			notify({ 
				type: 'error', 
				text: 'Failed to run Matrix validation. Please check your connection.' 
			});
		} finally {
			validationLoading = false;
		}
	}

	function toggleDetails(check: string) {
		showDetails[check] = !showDetails[check];
	}

	function getStatusIcon(status: string) {
		switch (status) {
			case 'success': return '✅';
			case 'warning': return '⚠️';
			case 'error': return '❌';
			default: return '❓';
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'success': return 'var(--success-color, #22c55e)';
			case 'warning': return 'var(--warning-color, #f59e0b)';
			case 'error': return 'var(--error-color, #ef4444)';
			default: return 'var(--text-color)';
		}
	}

	// Run validation when component loads
	runValidation();
</script>

<div class="modal-backdrop" onclick={onClose}>
	<div class="modal" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<h2>Matrix Setup Validation</h2>
			<button class="close-button" onclick={onClose}>×</button>
		</div>

		<div class="modal-content">
			<div class="validation-intro">
				<p>This tool performs comprehensive validation of your Matrix/Dendrite server configuration to ensure everything is working properly for Movie Club community features.</p>
			</div>

			{#if validationLoading}
				<div class="loading-section">
					<Spinner />
					<span>Running Matrix validation checks...</span>
				</div>
			{:else if validationResults}
				<div class="validation-results">
					<div class="overall-status" style="border-left: 4px solid {getStatusColor(validationResults.overallStatus)}">
						<div class="status-header">
							<span class="status-icon">{getStatusIcon(validationResults.overallStatus)}</span>
							<h3>Overall Status: {validationResults.overallStatus.toUpperCase()}</h3>
						</div>
						<p>{validationResults.summary}</p>
					</div>

					<div class="checks-list">
						<h4>Validation Checks</h4>
						{#each validationResults.results as result}
							<div class="check-item" style="border-left: 3px solid {getStatusColor(result.status)}">
								<div class="check-header" onclick={() => toggleDetails(result.check)}>
									<div class="check-info">
										<span class="check-icon">{getStatusIcon(result.status)}</span>
										<span class="check-name">{result.check}</span>
									</div>
									<div class="check-message">{result.message}</div>
									{#if result.details}
										<button class="details-toggle" type="button">
											{showDetails[result.check] ? '▼' : '▶'} Details
										</button>
									{/if}
								</div>
								
								{#if result.details && showDetails[result.check]}
									<div class="check-details">
										<div class="details-content">
											{result.details}
										</div>
									</div>
								{/if}
							</div>
						{/each}
					</div>

					{#if validationResults.overallStatus === 'success'}
						<div class="success-actions">
							<p><strong>Great!</strong> Your Matrix setup is fully configured and ready to use.</p>
						</div>
					{:else if validationResults.overallStatus === 'warning'}
						<div class="warning-actions">
							<p><strong>Almost there!</strong> Your Matrix setup is functional but has some optional configurations missing.</p>
						</div>
					{:else}
						<div class="error-actions">
							<p><strong>Action required:</strong> Please resolve the errors above before using Matrix features.</p>
							<button onclick={() => { troubleshootingModalOpen = true; }} class="troubleshooting-button">
								📋 View Troubleshooting Guide
							</button>
						</div>
					{/if}
				</div>
			{/if}
		</div>

		<div class="modal-footer">
			<div class="footer-actions">
				<button onclick={() => { troubleshootingModalOpen = true; }} class="help">
					📋 Troubleshooting Guide
				</button>
				<button onclick={runValidation} disabled={validationLoading} class="secondary">
					{#if validationLoading}
						<Spinner />
					{:else}
						Rerun Validation
					{/if}
				</button>
				<button onclick={onClose} class="primary">Close</button>
			</div>
		</div>
	</div>
</div>

{#if troubleshootingModalOpen}
	<MatrixTroubleshooting onClose={() => { troubleshootingModalOpen = false; }} />
{/if}

<style lang="scss">
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal {
		background: var(--bg-color);
		border-radius: 12px;
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
		width: 90%;
		max-width: 800px;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 20px;
		border-bottom: 1px solid var(--accent-color);

		h2 {
			margin: 0;
			font-size: 1.5em;
		}

		.close-button {
			background: none;
			border: none;
			font-size: 24px;
			cursor: pointer;
			color: var(--text-color);
			padding: 5px;
			border-radius: 50%;
			width: 35px;
			height: 35px;
			display: flex;
			align-items: center;
			justify-content: center;

			&:hover {
				background: var(--accent-color);
			}
		}
	}

	.modal-content {
		flex: 1;
		overflow-y: auto;
		padding: 20px;
	}

	.validation-intro {
		margin-bottom: 25px;
		padding: 15px;
		background: var(--bg-color);
		border-radius: 8px;
		border-left: 4px solid var(--accent-color);

		p {
			margin: 0;
			line-height: 1.6;
		}
	}

	.loading-section {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 15px;
		padding: 40px;

		span {
			font-size: 1.1em;
		}
	}

	.validation-results {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.overall-status {
		padding: 20px;
		background: var(--bg-color);
		border-radius: 8px;

		.status-header {
			display: flex;
			align-items: center;
			gap: 10px;
			margin-bottom: 10px;

			.status-icon {
				font-size: 1.5em;
			}

			h3 {
				margin: 0;
				font-size: 1.2em;
			}
		}

		p {
			margin: 0;
			font-size: 1.05em;
			line-height: 1.5;
		}
	}

	.checks-list {
		h4 {
			margin: 0 0 15px 0;
			font-size: 1.1em;
			color: var(--text-color);
		}
	}

	.check-item {
		background: var(--bg-color);
		border-radius: 8px;
		margin-bottom: 10px;
		overflow: hidden;
		border: 1px solid var(--accent-color);
	}

	.check-header {
		padding: 15px;
		cursor: pointer;
		transition: background-color 0.2s ease;

		&:hover {
			background: var(--bg-color-accent);
		}

		.check-info {
			display: flex;
			align-items: center;
			gap: 10px;
			margin-bottom: 5px;

			.check-icon {
				font-size: 1.2em;
			}

			.check-name {
				font-weight: 500;
				font-size: 1.05em;
			}
		}

		.check-message {
			color: var(--text-color-accent);
			margin-bottom: 8px;
			line-height: 1.4;
		}

		.details-toggle {
			background: none;
			border: none;
			color: var(--accent-color);
			cursor: pointer;
			font-size: 0.9em;
			padding: 4px 8px;
			border-radius: 4px;
			transition: background-color 0.2s ease;

			&:hover {
				background: var(--accent-color-light, rgba(59, 130, 246, 0.1));
			}
		}
	}

	.check-details {
		border-top: 1px solid var(--accent-color);
		background: var(--bg-color);
	}

	.details-content {
		padding: 15px;
		font-family: monospace;
		font-size: 0.9em;
		color: var(--text-color-accent);
		line-height: 1.4;
		word-break: break-all;
	}

	.success-actions, .warning-actions, .error-actions {
		padding: 15px;
		border-radius: 8px;
		margin-top: 10px;

		p {
			margin: 0;
			line-height: 1.5;
		}
	}

	.success-actions {
		background: rgba(34, 197, 94, 0.1);
		border: 1px solid rgba(34, 197, 94, 0.3);
		color: var(--success-color, #22c55e);
	}

	.warning-actions {
		background: rgba(245, 158, 11, 0.1);
		border: 1px solid rgba(245, 158, 11, 0.3);
		color: var(--warning-color, #f59e0b);
	}

	.error-actions {
		background: rgba(239, 68, 68, 0.1);
		border: 1px solid rgba(239, 68, 68, 0.3);
		color: var(--error-color, #ef4444);

		.troubleshooting-button {
			margin-top: 10px;
			background: var(--error-color, #ef4444);
			color: white;
			border: none;
			padding: 8px 16px;
			border-radius: 6px;
			cursor: pointer;
			font-size: 0.9em;
			transition: all 0.2s ease;

			&:hover {
				background: var(--error-color-dark, #dc2626);
			}
		}
	}

	.modal-footer {
		padding: 20px;
		border-top: 1px solid var(--accent-color);
		background: var(--bg-color);
	}

	.footer-actions {
		display: flex;
		gap: 10px;
		justify-content: flex-end;

		button {
			padding: 10px 20px;
			border-radius: 6px;
			font-weight: 500;
			cursor: pointer;
			transition: all 0.2s ease;
			border: none;
			display: flex;
			align-items: center;
			gap: 8px;

			&.primary {
				background: var(--accent-color);
				color: white;

				&:hover {
					background: var(--accent-color-dark);
				}
			}

			&.secondary {
				background: transparent;
				color: var(--accent-color);
				border: 1px solid var(--accent-color);

				&:hover {
					background: var(--accent-color);
					color: white;
				}
			}

			&.help {
				background: transparent;
				color: var(--warning-color, #f59e0b);
				border: 1px solid var(--warning-color, #f59e0b);

				&:hover {
					background: var(--warning-color, #f59e0b);
					color: white;
				}
			}

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
			}
		}
	}

	@media (max-width: 768px) {
		.modal {
			width: 95%;
			max-height: 95vh;
		}

		.modal-content {
			padding: 15px;
		}

		.modal-header {
			padding: 15px;
		}

		.footer-actions {
			flex-direction: column;

			button {
				width: 100%;
				justify-content: center;
			}
		}
	}
</style>