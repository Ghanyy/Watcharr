<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import type { MovieClubCycleResponse, Content } from "@/types";
	import { nominateMovie, removeNomination } from "@/lib/util/api";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	// Modal imports moved to parent component

	// Component for collapsible reason text
	function CollapsibleReason(reason: string, maxLength: number = 40) {
		return {
			shouldCollapse: reason.length > maxLength,
			truncated: reason.substring(0, maxLength) + "...",
			full: reason,
		};
	}

	export let cycleData: MovieClubCycleResponse;

	const dispatch = createEventDispatcher<{
		nominationChanged: void;
		openSearchModal: void;
		nominateMovie: { content: Content; reason: string };
	}>();

	// Remove local modal state - parent will handle it
	let submitting = false;
	let expandedReasons: Set<string> = new Set(); // Track expanded reasons by unique ID

	function handleOpenSearchModal() {
		dispatch("openSearchModal");
	}

	function handleNominateMovie(content: Content, reason: string) {
		dispatch("nominateMovie", {
			content,
			reason,
		});
	}

	async function handleRemoveNomination(nominationId: number) {
		if (submitting) return;

		submitting = true;
		const success = await removeNomination(nominationId);

		if (success) {
			dispatch("nominationChanged");
		}
		submitting = false;
	}

	function toggleReasonExpansion(reasonId: string) {
		if (expandedReasons.has(reasonId)) {
			expandedReasons.delete(reasonId);
		} else {
			expandedReasons.add(reasonId);
		}
		expandedReasons = expandedReasons; // Trigger reactivity
	}

	$: userNominations = cycleData.userNominations;
	$: canNominate = cycleData.canNominate;
	// Calculate max nominations based on existing data
	// TODO: This should come from movie club settings API
	$: maxNominations =
		cycleData.maxNominations ||
		Math.max(userNominations.length + (canNominate ? 1 : 0), 1);
	// Local validation - check if user has reached the limit
	$: canNominateLocal =
		canNominate && userNominations.length < maxNominations && !submitting;
</script>

<div class="nominations-section">
	<div class="section-header">
		<h3>
			<Icon icon="add" />
			Movie Nominations
		</h3>
		<p>Nominate movies you'd like the group to consider for this cycle.</p>
	</div>

	<!-- Your Nominations -->
	<div class="your-nominations">
		<h4>Your Nominations ({userNominations.length}/{maxNominations})</h4>

		{#if userNominations.length === 0}
			<div class="empty-state">
				<Icon icon="film" />
				<p>You haven't nominated any movies yet.</p>
				{#if canNominateLocal}
					<button class="nominate-btn" on:click={handleOpenSearchModal}>
						<Icon icon="add" />
						Nominate a Movie
					</button>
				{:else}
					<p class="limit-reached">Nomination limit reached</p>
				{/if}
			</div>
		{:else}
			<div class="nominations-grid">
				{#each userNominations as nomination}
					<div class="nomination-card">
						<div class="poster-container">
							<Poster
								media={nomination.content}
								showRating={false}
								disableInteraction={true}
							/>
						</div>
						<div class="nomination-details">
							<h5>{nomination.content?.title}</h5>
							{#if nomination.reason}
								{@const reasonData = CollapsibleReason(nomination.reason)}
								{@const reasonId = `user-${nomination.id}`}
								{@const isExpanded = expandedReasons.has(reasonId)}

								<div class="reason-container">
									<p class="reason">
										"{isExpanded
											? reasonData.full
											: reasonData.shouldCollapse
												? reasonData.truncated
												: reasonData.full}"
									</p>
									{#if reasonData.shouldCollapse}
										<button
											class="expand-btn"
											on:click={() => toggleReasonExpansion(reasonId)}
											type="button"
										>
											{isExpanded ? "Show less" : "Show more"}
										</button>
									{/if}
								</div>
							{/if}
							<button
								class="remove-btn"
								on:click={() => handleRemoveNomination(nomination.id)}
								disabled={submitting}
							>
								<Icon icon="trash" />
								Remove
							</button>
						</div>
					</div>
				{/each}
			</div>

			{#if canNominateLocal}
				<button class="nominate-btn" on:click={handleOpenSearchModal}>
					<Icon icon="add" />
					Nominate Another Movie
				</button>
			{/if}
		{/if}
	</div>

	<!-- All Nominations -->
	{#if cycleData.cycle.nominations && cycleData.cycle.nominations.length > 0}
		<div class="all-nominations">
			<h4>All Nominations ({cycleData.cycle.nominations.length})</h4>
			<div class="nominations-grid">
				{#each cycleData.cycle.nominations as nomination}
					<div class="nomination-card">
						<div class="poster-container">
							<Poster
								media={nomination.content}
								showRating={false}
								disableInteraction={true}
							/>
						</div>
						<div class="nomination-details">
							<h5>{nomination.content?.title}</h5>
							<div class="nominators">
								<p class="nominator-label">
									Nominated by:
									{#each nomination.nominators as nominator, i}
										<span class="nominator">{nominator.username}</span
										>{#if i < nomination.nominators.length - 1},
										{/if}
									{/each}
								</p>
							</div>
							{#if nomination.reasons && nomination.reasons.length > 0}
								<div class="reasons">
									{#each nomination.reasons as reason, i}
										{@const reasonData = CollapsibleReason(reason)}
										{@const reasonId = `all-${nomination.content.tmdbId}-${i}`}
										{@const isExpanded = expandedReasons.has(reasonId)}

										<div class="reason-container">
											<p class="reason">
												"{isExpanded
													? reasonData.full
													: reasonData.shouldCollapse
														? reasonData.truncated
														: reasonData.full}"
											</p>
											{#if reasonData.shouldCollapse}
												<button
													class="expand-btn"
													on:click={() => toggleReasonExpansion(reasonId)}
													type="button"
												>
													{isExpanded ? "Show less" : "Show more"}
												</button>
											{/if}
										</div>
									{/each}
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<!-- Modal moved to parent component to avoid transform issues -->

<style lang="scss">
	.nominations-section {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);

		// CSS custom properties for consistent design system
		--space-xs: 0.25rem;
		--space-sm: 0.5rem;
		--space-md: 1rem;
		--space-lg: 1.5rem;
		--space-xl: 2rem;
		--space-2xl: 3rem;

		--radius-sm: 4px;
		--radius-md: 8px;
		--radius-lg: 12px;
		--radius-xl: 16px;
		--radius-full: 50%;

		--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
		--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1),
			0 2px 4px -1px rgba(0, 0, 0, 0.06);
		--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1),
			0 4px 6px -2px rgba(0, 0, 0, 0.05);

		@media (prefers-color-scheme: dark) {
			--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.3);
			--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.4),
				0 2px 4px -1px rgba(0, 0, 0, 0.3);
			--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.5),
				0 4px 6px -2px rgba(0, 0, 0, 0.4);
		}
	}

	.section-header {
		text-align: center;
		margin-bottom: var(--space-lg);

		h3 {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: var(--space-sm);
			margin: 0 0 var(--space-sm) 0;
			font-size: 1.5rem;
			font-weight: 700;
			color: var(--text);
			line-height: 1.2;
		}

		p {
			color: var(--text-muted);
			margin: 0;
			line-height: 1.4;
			font-size: 0.95rem;
		}
	}

	.your-nominations,
	.all-nominations {
		h4 {
			margin: 0 0 var(--space-md) 0;
			padding-bottom: var(--space-sm);
			border-bottom: 1px solid var(--border);
			font-weight: 600;
			color: var(--text);
			font-size: 1.125rem;
		}
	}

	.empty-state {
		text-align: center;
		padding: var(--space-2xl) var(--space-md);
		color: var(--text-muted);
		border-radius: var(--radius-lg);
		border: 2px dashed var(--border);
		position: relative;
		transition: all 0.2s ease;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: var(--text-muted);
			opacity: 0.02;
			border-radius: inherit;
			pointer-events: none;
		}

		&:hover {
			&::before {
				background: var(--primary);
				opacity: 0.03;
			}
		}

		:global(svg) {
			font-size: 3rem;
			margin-bottom: var(--space-md);
			opacity: 0.4;
			color: var(--text-muted);
		}

		p {
			margin: 0 0 var(--space-lg) 0;
			line-height: 1.4;
			position: relative;
			z-index: 1;

			&:last-child {
				margin-bottom: 0;
			}
		}

		.limit-reached {
			color: var(--warning, #f59e0b);
			font-weight: 600;
			background: var(--background);
			padding: var(--space-sm) var(--space-md);
			border-radius: var(--radius-md);
			border: 1px solid var(--warning, #f59e0b);
			margin-top: var(--space-sm);
			position: relative;
			z-index: 1;
		}
	}

	.nominations-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
	}

	.nomination-card {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		overflow: hidden;
		transition: all 0.3s ease;
		box-shadow: var(--shadow-sm);

		&:hover {
			transform: translateY(-2px);
			box-shadow: var(--shadow-lg);
		}

		.poster-container {
			aspect-ratio: 2/3;
			overflow: hidden;
			display: flex;
			align-items: center;
			justify-content: center;
			border: 2px solid transparent;
			border-radius: var(--radius-md);
			transition: border-color 0.3s ease;

			&:hover {
				border-color: var(--primary);
			}

			:global(li) {
				list-style: none;
				margin: 0;
				padding: 0;
				width: 100%;
				height: 100%;
				display: flex;
				align-items: center;
				justify-content: center;
				cursor: default !important;
			}

			:global(.container) {
				width: 100% !important;
				height: 100% !important;
				display: flex !important;
				align-items: center !important;
				justify-content: center !important;
				transform: none !important;
				transition: none !important;
			}

			:global(.active .container) {
				transform: none !important;
			}

			:global(.inner) {
				opacity: 0 !important;
				pointer-events: none !important;
			}

			:global(img) {
				max-width: 100%;
				max-height: 100%;
				object-fit: cover;
				object-position: center;
				filter: none !important;
				mix-blend-mode: normal !important;
			}
		}

		.nomination-details {
			padding: var(--space-md);

			h5 {
				margin: 0 0 var(--space-sm) 0;
				font-size: 0.95rem;
				font-weight: 600;
				line-height: 1.2;
				color: var(--text);
			}

			.nominators {
				margin: 0 0 var(--space-sm) 0;
			}

			.nominator-label {
				font-size: 0.8rem;
				color: var(--text-muted);
				margin: 0;
				font-weight: 500;
				line-height: 1.4;
			}

			.nominator {
				color: var(--primary);
				font-weight: 600;
			}

			.reasons {
				margin: 0 0 var(--space-md) 0;
			}

			.reason-container {
				margin-bottom: var(--space-sm);
				padding: var(--space-sm);
				border-radius: var(--radius-md);
				border-left: 3px solid var(--primary);
				border: 1px solid var(--border);
				position: relative;

				&::before {
					content: "";
					position: absolute;
					inset: 0;
					background: var(--primary);
					opacity: 0.02;
					border-radius: inherit;
					pointer-events: none;
				}
			}

			.reason-container:last-child {
				margin-bottom: var(--space-md);
			}

			.reason {
				font-size: 0.85rem;
				color: var(--text-muted);
				font-style: italic;
				margin: 0 0 var(--space-xs) 0;
				line-height: 1.4;
			}

			.expand-btn {
				background: none;
				border: none;
				color: var(--primary);
				font-size: 0.75rem;
				cursor: pointer;
				padding: var(--space-xs) 0;
				text-decoration: underline;
				font-family: inherit;
				font-weight: 500;
				transition: color 0.2s ease;

				&:hover {
					color: var(--primary-dark, var(--primary));
				}

				&:focus {
					outline: none;
					text-decoration: none;
					box-shadow: 0 0 0 2px var(--primary);
					border-radius: var(--radius-sm);
				}
			}

			.remove-btn {
				display: inline-flex;
				align-items: center;
				justify-content: center;
				gap: var(--space-xs);
				width: 100%;
				padding: var(--space-sm) var(--space-md);
				background: var(--danger, #dc3545);
				color: white;
				border: none;
				border-radius: var(--radius-md);
				font-size: 0.8rem;
				font-weight: 500;
				cursor: pointer;
				transition: all 0.2s ease;
				box-shadow: var(--shadow-sm);

				&:hover:not(:disabled) {
					background: var(--danger-dark, #c82333);
					transform: translateY(-1px);
					box-shadow: var(--shadow-md);
				}

				&:disabled {
					opacity: 0.6;
					cursor: not-allowed;
					transform: none;
					box-shadow: var(--shadow-sm);
				}

				&:focus {
					outline: none;
					box-shadow:
						var(--shadow-md),
						0 0 0 3px rgba(220, 53, 69, 0.2);
				}

				&:active:not(:disabled) {
					transform: translateY(0);
					box-shadow: var(--shadow-sm);
				}
			}
		}
	}

	.nominate-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
		padding: var(--space-sm) var(--space-lg);
		background: var(--primary);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		font-weight: 500;
		cursor: pointer;
		margin: 0 auto;
		transition: all 0.2s ease;
		box-shadow: var(--shadow-sm);
		min-width: 160px;

		&:hover:not(:disabled) {
			background: var(--primary-dark, var(--primary));
			transform: translateY(-1px);
			box-shadow: var(--shadow-md);
		}

		&:disabled {
			opacity: 0.6;
			cursor: not-allowed;
			transform: none;
			box-shadow: var(--shadow-sm);
		}

		&:focus {
			outline: none;
			box-shadow:
				var(--shadow-md),
				0 0 0 3px rgba(59, 130, 246, 0.2);
		}

		&:active:not(:disabled) {
			transform: translateY(0);
			box-shadow: var(--shadow-sm);
		}
	}

	@media (max-width: 768px) {
		.nominations-section {
			gap: var(--space-lg);
		}

		.section-header {
			margin-bottom: var(--space-md);

			h3 {
				font-size: 1.25rem;
			}

			p {
				font-size: 0.9rem;
			}
		}

		.nominations-grid {
			grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
			gap: var(--space-sm);
		}

		.nomination-card {
			.nomination-details {
				padding: var(--space-sm);

				h5 {
					font-size: 0.85rem;
				}

				.reason,
				.nominator-label {
					font-size: 0.75rem;
				}

				.reason-container {
					padding: var(--space-xs);
				}

				.remove-btn {
					padding: var(--space-xs) var(--space-sm);
					font-size: 0.75rem;
				}
			}
		}

		.nominate-btn {
			padding: var(--space-xs) var(--space-md);
			font-size: 0.9rem;
			min-width: 140px;
		}

		.empty-state {
			padding: var(--space-xl) var(--space-sm);

			:global(svg) {
				font-size: 2.5rem;
			}
		}
	}
</style>
