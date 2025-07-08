<script lang="ts">
	import Menu from "../Menu.svelte";
	import { goto } from "$app/navigation";

	interface Props {
		close: () => void;
	}

	let { close }: Props = $props();

	function navigateAndClose(path: string) {
		goto(path);
		close();
	}
</script>

<Menu conf={{ width: "140px", arrowRight: "96px" }}>
	<div class="list">
		<a href="/movie-club" onclick={() => navigateAndClose("/movie-club")}>
			Dashboard
		</a>
		<a
			href="/movie-club/community"
			onclick={() => navigateAndClose("/movie-club/community")}
		>
			Community
		</a>
	</div>
</Menu>

<style lang="scss">
	.list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	a {
		color: var(--text-100);
		text-decoration: none;
		padding: 8px 12px;
		border-radius: 4px;
		transition: background-color 0.2s ease;

		&:hover {
			background-color: var(--bg-200);
		}
	}
</style>
