<script lang="ts">
	import type { MovieClubCycle } from "@/types";
	import Icon from "@/lib/Icon.svelte";

	export let cycle: MovieClubCycle;

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString(undefined, {
			month: "short",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit"
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
			case "nomination": return "add";
			case "voting": return "check";
			case "watching": return "play";
			default: return "film";
		}
	}

	function getPhaseColor(phase: string): string {
		switch (phase) {
			case "nomination": return "#3b82f6"; // blue
			case "voting": return "#f59e0b"; // amber
			case "watching": return "#10b981"; // emerald
			default: return "#6b7280"; // gray
		}
	}

	$: phaseIcon = getPhaseIcon(cycle.phase);
	$: phaseColor = getPhaseColor(cycle.phase);
	$: timeRemaining = getTimeRemaining(cycle.phaseEndDate);
</script>

<div class="phase-status">
	<div class="phase-header">
		<div class="phase-info">
			<div class="phase-icon" style="background-color: {phaseColor}">
				<Icon icon={phaseIcon} />
			</div>
			<div class="phase-details">
				<h2>{cycle.name}</h2>
				<p class="phase-name">
					{cycle.phase.charAt(0).toUpperCase() + cycle.phase.slice(1)} Phase
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

	<div class="phase-progress">
		<div class="progress-steps">
			<div class="step" class:active={cycle.phase === "nomination"} class:completed={["voting", "watching"].includes(cycle.phase)}>
				<div class="step-icon">
					<Icon icon="add" />
				</div>
				<span>Nomination</span>
			</div>
			<div class="step" class:active={cycle.phase === "voting"} class:completed={cycle.phase === "watching"}>
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
</div>

<style lang="scss">
	.phase-status {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
	}

	.phase-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1rem;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.phase-info {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.phase-icon {
		width: 50px;
		height: 50px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		font-size: 1.25rem;
	}

	.phase-details {
		h2 {
			margin: 0 0 0.25rem 0;
			font-size: 1.5rem;
			color: var(--text);
		}

		.phase-name {
			margin: 0;
			color: var(--text-muted);
			font-weight: 500;
		}
	}

	.phase-timing {
		text-align: right;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.time-remaining {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-weight: 600;
		color: var(--primary);
	}

	.phase-end {
		small {
			color: var(--text-muted);
		}
	}

	.cycle-description {
		margin: 1rem 0;
		color: var(--text-muted);
		font-style: italic;
	}

	.phase-progress {
		margin-top: 1.5rem;
		padding-top: 1.5rem;
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
		gap: 0.5rem;
		position: relative;
		z-index: 2;

		.step-icon {
			width: 50px;
			height: 50px;
			border-radius: 50%;
			border: 3px solid var(--border);
			background: var(--background);
			display: flex;
			align-items: center;
			justify-content: center;
			transition: all 0.3s ease;
		}

		span {
			font-size: 0.875rem;
			color: var(--text-muted);
			font-weight: 500;
		}

		&.active {
			.step-icon {
				border-color: var(--primary);
				background: var(--primary);
				color: white;
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
			}

			span {
				color: var(--success);
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