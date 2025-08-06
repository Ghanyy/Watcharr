<script lang="ts">
	import type { MovieClubCycle } from "@/types";
	import Icon from "@/lib/Icon.svelte";

	export let cycle: MovieClubCycle;

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString(undefined, {
			month: "short",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit",
		});
	}

	function getTimeRemaining(endDate: string): string {
		const now = new Date();
		const end = new Date(endDate);
		const diff = end.getTime() - now.getTime();

		if (diff <= 0) return "Phase ended";

		const days = Math.floor(diff / (1000 * 60 * 60 * 24));
		const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
		const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

		if (days > 0) return `${days}d ${hours}h remaining`;
		if (hours > 0) return `${hours}h ${minutes}m remaining`;
		return `${minutes}m remaining`;
	}

	function getPhaseIcon(phase: string): string {
		switch (phase) {
			case "nomination":
				return "add";
			case "voting":
				return "check";
			case "watching":
				return "play";
			default:
				return "film";
		}
	}

	function getPhaseColorClass(phase: string): string {
		switch (phase) {
			case "nomination":
				return "phase-nomination";
			case "voting":
				return "phase-voting";
			case "watching":
				return "phase-watching";
			default:
				return "phase-default";
		}
	}

	$: phaseIcon = getPhaseIcon(cycle.phase);
	$: phaseColorClass = getPhaseColorClass(cycle.phase);
	$: timeRemaining = getTimeRemaining(cycle.phaseEndDate);
</script>

<div class="phase-status">
	<div class="phase-header">
		<div class="phase-info">
			<div class="phase-icon {phaseColorClass}">
				<Icon icon={phaseIcon} />
			</div>
			<div class="phase-details">
				<div class="title-row">
					<h2>{cycle.name}</h2>
					{#if cycle.isAdHoc}
						<span class="adhoc-badge">
							<Icon icon="sparkles" />
							Ad-hoc
						</span>
					{/if}
				</div>
				<p class="phase-name">
					{#if cycle.isAdHoc}
						Ad-hoc Session • {cycle.phase.charAt(0).toUpperCase() +
							cycle.phase.slice(1)} Phase
					{:else}
						{cycle.phase.charAt(0).toUpperCase() + cycle.phase.slice(1)} Phase
					{/if}
				</p>
			</div>
		</div>
		<div class="phase-timing">
			<div class="time-remaining">
				<Icon icon="clock" />
				<span>{timeRemaining}</span>
			</div>
			<div class="phase-end">
				<small>Ends {formatDate(cycle.phaseEndDate)}</small>
			</div>
		</div>
	</div>

	{#if cycle.description}
		<p class="cycle-description">{cycle.description}</p>
	{/if}

	{#if !cycle.isAdHoc}
		<div class="phase-progress">
			<div class="progress-steps">
				<div
					class="step"
					class:active={cycle.phase === "nomination"}
					class:completed={["voting", "watching"].includes(cycle.phase)}
				>
					<div class="step-icon">
						<Icon icon="add" />
					</div>
					<span>Nomination</span>
				</div>
				<div
					class="step"
					class:active={cycle.phase === "voting"}
					class:completed={cycle.phase === "watching"}
				>
					<div class="step-icon">
						<Icon icon="check" />
					</div>
					<span>Voting</span>
				</div>
				<div class="step" class:active={cycle.phase === "watching"}>
					<div class="step-icon">
						<Icon icon="play" />
					</div>
					<span>Watching</span>
				</div>
			</div>
		</div>
	{:else}
		<div class="phase-progress adhoc-progress">
			<div class="progress-steps single-step">
				<div class="step active adhoc-step">
					<div class="step-icon">
						<Icon icon="sparkles" />
					</div>
					<span>Instant Watch Session</span>
				</div>
			</div>
			<div class="adhoc-info">
				<small>
					<Icon icon="clock" />
					{cycle.adHocDurationHours}h duration
				</small>
			</div>
		</div>
	{/if}
</div>

<style lang="scss">
	.phase-status {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		box-shadow: var(--shadow-md);

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

	.phase-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-md);
		flex-wrap: wrap;
		gap: var(--space-md);
	}

	.phase-info {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.phase-icon {
		width: 50px;
		height: 50px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		font-size: 1.25rem;
		transition: transform 0.2s ease;

		&:hover {
			transform: scale(1.05);
		}

		&.phase-nomination {
			background: var(--info, #3b82f6);
		}

		&.phase-voting {
			background: var(--warning, #f59e0b);
		}

		&.phase-watching {
			background: var(--success, #10b981);
		}

		&.phase-default {
			background: var(--text-muted, #6b7280);
		}
	}

	.phase-details {
		h2 {
			margin: 0 0 var(--space-xs) 0;
			font-size: 1.5rem;
			color: var(--text);
			font-weight: 700;
			line-height: 1.2;
		}

		.phase-name {
			margin: 0;
			color: var(--text-muted);
			font-weight: 500;
			font-size: 0.875rem;
			text-transform: uppercase;
			letter-spacing: 0.025em;
		}
	}

	.phase-timing {
		text-align: right;
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.time-remaining {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		font-weight: 600;
		color: var(--primary);
		font-size: 0.9rem;
	}

	.phase-end {
		small {
			color: var(--text-muted);
		}
	}

	.cycle-description {
		margin: var(--space-md) 0;
		color: var(--text-muted);
		font-style: italic;
		line-height: 1.5;
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-md);
		border-left: 3px solid var(--primary);
		position: relative;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: var(--primary);
			opacity: 0.03;
			border-radius: inherit;
			pointer-events: none;
		}
	}

	.phase-progress {
		margin-top: var(--space-lg);
		padding-top: var(--space-lg);
		border-top: 1px solid var(--border);
	}

	.progress-steps {
		display: flex;
		justify-content: space-between;
		position: relative;

		&::before {
			content: "";
			position: absolute;
			top: 25px;
			left: 25px;
			right: 25px;
			height: 2px;
			background: var(--border);
			z-index: 1;
		}
	}

	.step {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-sm);
		position: relative;
		z-index: 2;

		.step-icon {
			width: 50px;
			height: 50px;
			border-radius: var(--radius-full);
			border: 3px solid var(--border);
			background: var(--background);
			display: flex;
			align-items: center;
			justify-content: center;
			transition: all 0.3s ease;
			box-shadow: var(--shadow-sm);
		}

		span {
			font-size: 0.875rem;
			color: var(--text-muted);
			font-weight: 500;
			text-align: center;
			transition: all 0.3s ease;
		}

		&.active {
			.step-icon {
				border-color: var(--primary);
				background: var(--primary);
				color: white;
				box-shadow: var(--shadow-md);
				transform: scale(1.05);
			}

			span {
				color: var(--primary);
				font-weight: 600;
			}
		}

		&.completed {
			.step-icon {
				border-color: var(--success);
				background: var(--success);
				color: white;
				box-shadow: var(--shadow-md);
			}

			span {
				color: var(--success);
				font-weight: 600;
			}
		}
	}

	// Ad-hoc cycle styles
	.title-row {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		flex-wrap: wrap;
	}

	.adhoc-badge {
		display: inline-flex;
		align-items: center;
		gap: var(--space-xs);
		padding: var(--space-xs) var(--space-sm);
		background: linear-gradient(135deg, #8b5cf6, #a855f7);
		color: white;
		border-radius: var(--radius-full);
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.025em;
		box-shadow: var(--shadow-sm);
		animation: shimmer 2s ease-in-out infinite alternate;

		:global(svg) {
			width: 12px;
			height: 12px;
		}
	}

	@keyframes shimmer {
		0% {
			opacity: 0.8;
		}
		100% {
			opacity: 1;
		}
	}

	.adhoc-progress {
		.single-step {
			justify-content: center;
		}

		.adhoc-step {
			background: linear-gradient(135deg, #8b5cf6, #a855f7);
			color: white;
			border-color: #7c3aed;
			box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);

			.step-icon {
				background: rgba(255, 255, 255, 0.2);
			}
		}

		.adhoc-info {
			display: flex;
			justify-content: center;
			margin-top: var(--space-sm);
			color: var(--text-muted);

			small {
				display: flex;
				align-items: center;
				gap: var(--space-xs);
				font-size: 0.8rem;
			}

			:global(svg) {
				width: 14px;
				height: 14px;
			}
		}
	}

	@media (max-width: 768px) {
		.phase-header {
			flex-direction: column;
			text-align: center;
		}

		.phase-timing {
			text-align: center;
		}

		.progress-steps {
			gap: 1rem;

			.step span {
				font-size: 0.75rem;
			}
		}
	}
</style>
