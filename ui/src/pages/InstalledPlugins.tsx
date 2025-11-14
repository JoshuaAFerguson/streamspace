import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  CardActions,
  Button,
  IconButton,
  Chip,
  CircularProgress,
  Alert,
  Switch,
  FormControlLabel,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Tooltip,
} from '@mui/material';
import {
  Settings as SettingsIcon,
  Delete as DeleteIcon,
  Info as InfoIcon,
  Extension as ExtensionIcon,
  Webhook as WebhookIcon,
  Api as ApiIcon,
  Dashboard as UiIcon,
  Palette as ThemeIcon,
} from '@mui/icons-material';
import Layout from '../components/Layout';
import { api, type InstalledPlugin } from '../lib/api';
import { toast } from '../lib/toast';

const pluginTypeIcons: Record<string, JSX.Element> = {
  extension: <ExtensionIcon fontSize="small" />,
  webhook: <WebhookIcon fontSize="small" />,
  api: <ApiIcon fontSize="small" />,
  ui: <UiIcon fontSize="small" />,
  theme: <ThemeIcon fontSize="small" />,
};

const pluginTypeColors: Record<string, string> = {
  extension: '#4CAF50',
  webhook: '#FF9800',
  api: '#2196F3',
  ui: '#9C27B0',
  theme: '#E91E63',
};

export default function InstalledPlugins() {
  const [loading, setLoading] = useState(true);
  const [plugins, setPlugins] = useState<InstalledPlugin[]>([]);
  const [filter, setFilter] = useState<'all' | 'enabled' | 'disabled'>('all');
  const [configDialogOpen, setConfigDialogOpen] = useState(false);
  const [selectedPlugin, setSelectedPlugin] = useState<InstalledPlugin | null>(null);
  const [configJson, setConfigJson] = useState('');

  useEffect(() => {
    loadPlugins();
  }, []);

  const loadPlugins = async () => {
    setLoading(true);
    try {
      const data = await api.listInstalledPlugins();
      setPlugins(data);
    } catch (error) {
      console.error('Failed to load installed plugins:', error);
      toast.error('Failed to load installed plugins');
    } finally {
      setLoading(false);
    }
  };

  const handleTogglePlugin = async (plugin: InstalledPlugin) => {
    try {
      if (plugin.enabled) {
        await api.disablePlugin(plugin.id);
        toast.success(`${plugin.displayName || plugin.name} disabled`);
      } else {
        await api.enablePlugin(plugin.id);
        toast.success(`${plugin.displayName || plugin.name} enabled`);
      }
      await loadPlugins();
    } catch (error) {
      console.error('Failed to toggle plugin:', error);
      toast.error('Failed to toggle plugin');
    }
  };

  const handleOpenConfig = (plugin: InstalledPlugin) => {
    setSelectedPlugin(plugin);
    setConfigJson(JSON.stringify(plugin.config || {}, null, 2));
    setConfigDialogOpen(true);
  };

  const handleSaveConfig = async () => {
    if (!selectedPlugin) return;

    try {
      const config = JSON.parse(configJson);
      await api.updatePluginConfig(selectedPlugin.id, config);
      toast.success('Configuration updated');
      setConfigDialogOpen(false);
      await loadPlugins();
    } catch (error) {
      console.error('Failed to update configuration:', error);
      toast.error('Invalid JSON or failed to update configuration');
    }
  };

  const handleUninstall = async (plugin: InstalledPlugin) => {
    if (!confirm(`Are you sure you want to uninstall ${plugin.displayName || plugin.name}?`)) {
      return;
    }

    try {
      await api.uninstallPlugin(plugin.id);
      toast.success(`${plugin.displayName || plugin.name} uninstalled`);
      await loadPlugins();
    } catch (error) {
      console.error('Failed to uninstall plugin:', error);
      toast.error('Failed to uninstall plugin');
    }
  };

  const filteredPlugins = plugins.filter(plugin => {
    if (filter === 'enabled') return plugin.enabled;
    if (filter === 'disabled') return !plugin.enabled;
    return true;
  });

  return (
    <Layout>
      <Box>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
          <Box>
            <Typography variant="h4" sx={{ fontWeight: 700 }}>
              Installed Plugins
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Manage your installed plugins
            </Typography>
          </Box>
        </Box>

        {/* Filter Tabs */}
        <Box mb={3} display="flex" gap={1}>
          <Chip
            label={`All (${plugins.length})`}
            onClick={() => setFilter('all')}
            color={filter === 'all' ? 'primary' : 'default'}
            variant={filter === 'all' ? 'filled' : 'outlined'}
          />
          <Chip
            label={`Enabled (${plugins.filter(p => p.enabled).length})`}
            onClick={() => setFilter('enabled')}
            color={filter === 'enabled' ? 'primary' : 'default'}
            variant={filter === 'enabled' ? 'filled' : 'outlined'}
          />
          <Chip
            label={`Disabled (${plugins.filter(p => !p.enabled).length})`}
            onClick={() => setFilter('disabled')}
            color={filter === 'disabled' ? 'primary' : 'default'}
            variant={filter === 'disabled' ? 'filled' : 'outlined'}
          />
        </Box>

        {/* Results */}
        {loading ? (
          <Box display="flex" justifyContent="center" py={8}>
            <CircularProgress />
          </Box>
        ) : filteredPlugins.length === 0 ? (
          <Alert severity="info">
            {filter === 'all'
              ? 'No plugins installed yet. Browse the catalog to find plugins!'
              : `No ${filter} plugins.`}
          </Alert>
        ) : (
          <Grid container spacing={3}>
            {filteredPlugins.map((plugin) => (
              <Grid item xs={12} sm={6} md={4} key={plugin.id}>
                <Card
                  sx={{
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                    opacity: plugin.enabled ? 1 : 0.6,
                  }}
                >
                  <CardContent sx={{ flexGrow: 1 }}>
                    <Box display="flex" justifyContent="space-between" alignItems="start" mb={2}>
                      <Box display="flex" alignItems="center" gap={1}>
                        {plugin.pluginType && (
                          <Box
                            sx={{
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              width: 32,
                              height: 32,
                              borderRadius: '50%',
                              bgcolor: pluginTypeColors[plugin.pluginType] || '#757575',
                              color: 'white',
                            }}
                          >
                            {pluginTypeIcons[plugin.pluginType] || <ExtensionIcon fontSize="small" />}
                          </Box>
                        )}
                        <Box>
                          <Typography variant="h6" sx={{ fontSize: '1rem', fontWeight: 600 }}>
                            {plugin.displayName || plugin.name}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            v{plugin.version}
                          </Typography>
                        </Box>
                      </Box>
                      <FormControlLabel
                        control={
                          <Switch
                            checked={plugin.enabled}
                            onChange={() => handleTogglePlugin(plugin)}
                            size="small"
                          />
                        }
                        label=""
                      />
                    </Box>

                    {plugin.description && (
                      <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{
                          mb: 2,
                          minHeight: 40,
                          display: '-webkit-box',
                          WebkitLineClamp: 2,
                          WebkitBoxOrient: 'vertical',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                        }}
                      >
                        {plugin.description}
                      </Typography>
                    )}

                    <Box display="flex" gap={0.5} flexWrap="wrap" mb={1}>
                      <Chip
                        label={plugin.enabled ? 'Enabled' : 'Disabled'}
                        size="small"
                        color={plugin.enabled ? 'success' : 'default'}
                      />
                      {plugin.pluginType && (
                        <Chip
                          label={plugin.pluginType}
                          size="small"
                          sx={{
                            bgcolor: pluginTypeColors[plugin.pluginType] + '20',
                            color: pluginTypeColors[plugin.pluginType],
                          }}
                        />
                      )}
                    </Box>

                    <Typography variant="caption" color="text.secondary" display="block" mt={1}>
                      Installed by {plugin.installedBy} on {new Date(plugin.installedAt).toLocaleDateString()}
                    </Typography>
                  </CardContent>

                  <CardActions sx={{ justifyContent: 'space-between', px: 2, pb: 2 }}>
                    <Box display="flex" gap={1}>
                      <Tooltip title="Configure">
                        <IconButton
                          size="small"
                          onClick={() => handleOpenConfig(plugin)}
                        >
                          <SettingsIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      {plugin.manifest && (
                        <Tooltip title="View Details">
                          <IconButton size="small">
                            <InfoIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      )}
                    </Box>
                    <Tooltip title="Uninstall">
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => handleUninstall(plugin)}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </CardActions>
                </Card>
              </Grid>
            ))}
          </Grid>
        )}

        {/* Configuration Dialog */}
        <Dialog open={configDialogOpen} onClose={() => setConfigDialogOpen(false)} maxWidth="md" fullWidth>
          <DialogTitle>
            Configure {selectedPlugin?.displayName || selectedPlugin?.name}
          </DialogTitle>
          <DialogContent>
            <Typography variant="body2" color="text.secondary" mb={2}>
              Edit the plugin configuration as JSON. Invalid JSON will not be saved.
            </Typography>
            <TextField
              fullWidth
              multiline
              rows={12}
              value={configJson}
              onChange={(e) => setConfigJson(e.target.value)}
              placeholder="{}"
              sx={{ fontFamily: 'monospace' }}
            />
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setConfigDialogOpen(false)}>Cancel</Button>
            <Button variant="contained" onClick={handleSaveConfig}>
              Save Configuration
            </Button>
          </DialogActions>
        </Dialog>
      </Box>
    </Layout>
  );
}
