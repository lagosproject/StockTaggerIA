local LrView = import 'LrView'
local LrPathUtils = import 'LrPathUtils'
local LrFileUtils = import 'LrFileUtils'
local LrErrors = import 'LrErrors'
local LrDialogs = import 'LrDialogs'
local LrTasks = import 'LrTasks'

local exportServiceProvider = {}

exportServiceProvider.exportPresetFields = {
    { key = 'platform', default = 'stock' },
    { key = 'lang', default = 'en' },
    { key = 'dryRun', default = false },
    { key = 'customCliPath', default = '' },
}

function exportServiceProvider.sectionsForTopOfDialog( f, propertyTable )
    return {
        {
            title = "StockTaggerIA - Stock SEO Metadata",
            synopsis = function()
                return string.format("Preset: %s | Lang: %s", propertyTable.platform or 'stock', propertyTable.lang or 'en')
            end,

            f:row {
                f:static_text {
                    title = "Platform Preset:",
                    alignment = 'right',
                    width = LrView.share 'label_width',
                },
                f:popup_menu {
                    value = LrView.bind 'platform',
                    items = {
                        { title = "Unsplash & Pexels (Max 30 tags)", value = "stock" },
                        { title = "Pixabay (Max 25 tags)", value = "pixabay" },
                    },
                },
            },

            f:row {
                f:static_text {
                    title = "Metadata Language:",
                    alignment = 'right',
                    width = LrView.share 'label_width',
                },
                f:popup_menu {
                    value = LrView.bind 'lang',
                    items = {
                        { title = "English (en)", value = "en" },
                        { title = "Spanish (es)", value = "es" },
                        { title = "French (fr)", value = "fr" },
                    },
                },
            },

            f:row {
                f:static_text {
                    title = "Mode:",
                    alignment = 'right',
                    width = LrView.share 'label_width',
                },
                f:checkbox {
                    title = "Dry Run (Simulate only, do not modify files)",
                    value = LrView.bind 'dryRun',
                },
            },

            f:row {
                f:static_text {
                    title = "Custom CLI Path:",
                    alignment = 'right',
                    width = LrView.share 'label_width',
                },
                f:edit_field {
                    value = LrView.bind 'customCliPath',
                    width_in_chars = 28,
                },
            },
        },
    }
end

function exportServiceProvider.postProcessRenderedPhotos( functionContext, exportContext )
    local propertyTable = exportContext.propertyTable
    local platform = propertyTable.platform or 'stock'
    local lang = propertyTable.lang or 'en'
    local dryRun = propertyTable.dryRun or false
    local cliPath = propertyTable.customCliPath

    -- Auto-locate stocktaggeria executable if not set
    if not cliPath or cliPath == '' then
        if WIN_ENV then
            local localExe = LrPathUtils.child( _PLUGIN.path, "stocktaggeria.exe" )
            if LrFileUtils.exists( localExe ) then
                cliPath = localExe
            else
                cliPath = "stocktaggeria.exe"
            end
        else
            local localBin = LrPathUtils.child( _PLUGIN.path, "stocktaggeria" )
            if LrFileUtils.exists( localBin ) then
                cliPath = localBin
            else
                cliPath = "stocktaggeria"
            end
        end
    end

    -- Collect all exported photo folders
    local exportedDirs = {}
    for i, rendition in exportContext:renditions{ stopIfCanceled = true } do
        local success, pathOrMessage = rendition:waitForRender()
        if success and pathOrMessage then
            local dir = LrPathUtils.parent( pathOrMessage )
            if dir and not exportedDirs[dir] then
                exportedDirs[dir] = true
            end
        end
    end

    -- Execute StockTaggerIA across each exported directory
    for dir, _ in pairs( exportedDirs ) do
        local cmd = string.format('"%s" -p %s -l %s -d "%s"', cliPath, platform, lang, dir)
        if dryRun then
            cmd = cmd .. " --dry-run"
        end
        LrTasks.execute( cmd )
    end
end

return exportServiceProvider
