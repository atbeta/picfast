<template>
	<n-config-provider :theme-overrides="themeOverrides">
		<n-message-provider>
			<n-dialog-provider>
				<router-view />
			</n-dialog-provider>
		</n-message-provider>
	</n-config-provider>
</template>

<script setup lang="ts">
import type { GlobalThemeOverrides } from 'naive-ui'
import { NConfigProvider, NMessageProvider, NDialogProvider } from 'naive-ui'

const themeOverrides: GlobalThemeOverrides = {
	common: {
		primaryColor: '#3b82f6',
		primaryColorHover: '#2563eb',
		primaryColorPressed: '#1d4ed8',
		primaryColorSuppl: '#eff6ff',
		fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
		fontSize: '14px',
		borderRadius: '8px',
		borderRadiusSmall: '6px',
		lineHeight: '1.5',
		textColorBase: '#1e293b',
		textColor2: '#475569',
		textColor3: '#94a3b8',
		borderColor: '#e2e8f0',
		dividerColor: '#f1f5f9',
		boxShadow1: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
		boxShadow2: '0 4px 6px -1px rgb(0 0 0 / 0.07), 0 2px 4px -2px rgb(0 0 0 / 0.05)',
		boxShadow3: '0 10px 15px -3px rgb(0 0 0 / 0.08), 0 4px 6px -4px rgb(0 0 0 / 0.04)',
		heightSmall: '28px',
		heightMedium: '34px',
		heightLarge: '40px',
	},
	Button: {
		borderRadiusMedium: '8px',
		borderRadiusSmall: '6px',
		fontWeight: '500',
		textColorGhost: '#475569',
		textColorText: '#3b82f6',
		borderHover: '#93c5fd',
		borderPressed: '#3b82f6',
		borderFocus: '#3b82f6',
		rippleColor: '#3b82f6',
		boxShadowFocus: '0 0 0 2px rgb(59 130 246 / 0.15)',
	},
	Input: {
		borderRadius: '8px',
		border: '#e2e8f0',
		borderHover: '#cbd5e1',
		borderFocus: '#3b82f6',
		boxShadowFocus: '0 0 0 2px rgb(59 130 246 / 0.12)',
		placeholderColor: '#94a3b8',
		lineHeight: '1.5',
		paddingMedium: '6px 12px',
		paddingLarge: '8px 14px',
	},
	Select: {
		borderRadius: '8px',
		boxShadowFocus: '0 0 0 2px rgb(59 130 246 / 0.12)',
	},
	Card: {
		borderRadius: '12px',
		borderColor: '#f1f5f9',
		boxShadow: '0 1px 3px 0 rgb(0 0 0 / 0.04), 0 1px 2px -1px rgb(0 0 0 / 0.04)',
		paddingMedium: '18px',
		paddingSmall: '14px',
	},
	Tag: {
		borderRadius: '6px',
		fontWeight: '500',
	},
	DataTable: {
		borderRadius: '12px',
		thColor: '#f8fafc',
		thTextColor: '#64748b',
		thFontWeight: '600',
		tdColor: '#ffffff',
		tdColorHover: '#f8fafc',
		borderColor: '#f1f5f9',
		tdPaddingMedium: '10px 16px',
		thPaddingMedium: '10px 16px',
	},
	Pagination: {
		borderRadius: '8px',
		itemSize: '32px',
		itemBorder: '#e2e8f0',
		itemBorderActive: '#3b82f6',
		itemColorActive: '#eff6ff',
		buttonBorder: '#e2e8f0',
	},
	Modal: {
		borderRadius: '16px',
		boxShadow: '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
	},
	Switch: {
		borderRadius: '12px',
		railColorActive: '#3b82f6',
	},
	Checkbox: {
		borderRadius: '4px',
		borderChecked: '#3b82f6',
		colorChecked: '#3b82f6',
		boxShadowFocus: '0 0 0 2px rgb(59 130 246 / 0.15)',
	},
	Dropdown: {
		borderRadius: '10px',
		boxShadow: '0 4px 12px -2px rgb(0 0 0 / 0.08), 0 2px 4px -2px rgb(0 0 0 / 0.04)',
		optionHeightMedium: '34px',
		optionIconSizeMedium: '16px',
		padding: '6px',
	},
	Progress: {
		borderRadius: '4px',
		fillColor: '#3b82f6',
		railColor: '#f1f5f9',
	},
	Tabs: {
		borderRadius: '8px',
		tabFontWeightActive: '600',
		barColor: '#3b82f6',
		tabTextColorLine: '#64748b',
	},
	Spin: {
		color: '#3b82f6',
		sizeSmall: '16px',
		sizeMedium: '24px',
		sizeLarge: '32px',
	},
	Menu: {
		borderRadius: '8px',
		itemTextColor: '#475569',
		itemTextColorActive: '#1e293b',
		itemTextColorActiveHover: '#1e293b',
		itemColorActive: '#eff6ff',
		itemColorActiveHover: '#eff6ff',
		arrowColor: '#94a3b8',
	},
}
</script>
