import { useState, useEffect, useMemo } from 'react';
import { motion } from 'framer-motion';
import {
  PieChart as PieIcon,
  TrendingUp,
  AlertTriangle,
  Sparkles,
  Wallet,
  Calendar,
  User,
  Mail,
  Clock
} from 'lucide-react';
import {
  PieChart, Pie, Cell, Tooltip, ResponsiveContainer,
  BarChart, Bar, XAxis, YAxis, CartesianGrid,
  AreaChart, Area,
} from 'recharts';
import Navbar from '@/components/Navbar';
import { useAuth } from '@/context/AuthContext';
import { api } from '@/api/client';
import type { AnalyticsData, InsightsData } from '@/types';
import styles from './Dashboard.module.css';

const CHART_COLORS = [
  'var(--chart-1)', 'var(--chart-2)', 'var(--chart-3)', 'var(--chart-4)',
  'var(--chart-5)', 'var(--chart-6)', 'var(--chart-7)', 'var(--chart-8)',
];

export default function Dashboard() {
  const { name, email } = useAuth();
  const displayName = name || email?.split('@')[0] || 'User';

  const [analytics, setAnalytics] = useState<AnalyticsData | null>(null);
  const [insights, setInsights] = useState<InsightsData | null>(null);
  const [loading, setLoading] = useState(true);

  // Added logic to calculate monthly spending dynamically
  const { thisMonthTotal, previousMonths } = useMemo(() => {
    if (!analytics?.dailySpending) return { thisMonthTotal: 0, previousMonths: [] };

    const current = new Date();
    const months: Array<{ id: string; label: string; year: number; month: number; total: number }> = [];
    
    // Generate the last 5 months (including current)
    for (let i = 0; i < 5; i++) {
      const d = new Date(current.getFullYear(), current.getMonth() - i, 1);
      months.push({
        id: `${d.getFullYear()}-${d.getMonth()}`,
        label: d.toLocaleString('default', { month: 'short' }) + ' ' + d.getFullYear(),
        year: d.getFullYear(),
        month: d.getMonth(),
        total: 0
      });
    }
    months.reverse(); // Chronological order

    analytics.dailySpending.forEach((item) => {
      const date = new Date(item.date);
      const yr = date.getFullYear();
      const mo = date.getMonth();
      
      const match = months.find(m => m.year === yr && m.month === mo);
      if (match) {
        match.total += item.amount;
      }
    });

    const thisMonth = months[months.length - 1];
    const past = months.slice(0, months.length - 1).filter(m => m.total > 0);

    return {
      thisMonthTotal: thisMonth.total || 0, // In case there are no receipts
      previousMonths: past
    };
  }, [analytics?.dailySpending]);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const [analyticsRes, insightsRes] = await Promise.allSettled([
          api.get<AnalyticsData>('/useranalytics'),
          api.get<InsightsData>('/userinsights'),
        ]);
        if (analyticsRes.status === 'fulfilled') setAnalytics(analyticsRes.value);
        if (insightsRes.status === 'fulfilled') setInsights(insightsRes.value);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  const hasData = analytics && Array.isArray(analytics.categoryBreakdown) && analytics.categoryBreakdown.length > 0;

  return (
    <div className={styles.page}>
      <Navbar />
      <main className={styles.main}>
        {/* ── Profile header ── */}
        <motion.div className={`${styles.profileCard} glass-panel`} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
          <div className={styles.avatar}>
            <User size={32} />
          </div>
          <div className={styles.profileInfo}>
            <h1>Welcome back, <span className="gradient-text">{displayName}</span></h1>
            <p><Mail size={14} /> {email}</p>
          </div>
          {hasData && analytics && (
            <motion.div className={styles.profileStat} initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} transition={{ delay: 0.2 }}>
              <Wallet size={24} />
              <div>
                <span className={styles.statLabel}>This Month's Spending</span>
                <span className={styles.statValue}>{analytics.currency} {thisMonthTotal.toFixed(2)}</span>
              </div>
            </motion.div>
          )}
        </motion.div>

        {loading && (
          <div className={styles.loadingState}>
            <div className={styles.spinner} />
            <p>Loading your amazing dashboard...</p>
          </div>
        )}

        {!loading && !hasData && (
          <motion.div className={styles.emptyState} initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }}>
            <img src="https://illustrations.popsy.co/violet/app-launch.svg" alt="No data yet" className={styles.emptyIllustration} />
            <h2>Start your journey</h2>
            <p>Upload your first receipt to unlock AI-powered insights and beautiful analytics.</p>
          </motion.div>
        )}

        {!loading && hasData && (
          <motion.div initial="hidden" animate="visible" variants={{ visible: { transition: { staggerChildren: 0.1 } } }}>
            
            {/* ── Recent Months Strip ── */}
            {previousMonths.length > 0 && (
              <motion.div className={styles.recentMonthsContainer} variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }}>
                 <h2 className={styles.recentMonthsHeader}><Calendar size={20} style={{ color: 'var(--accent)' }}/> Recent Months</h2>
                 <div className={styles.monthsGrid}>
                   {previousMonths.map((m, i) => (
                     <motion.div key={m.id} className={`glass-panel ${styles.monthCard}`} transition={{ delay: i * 0.05 }} whileHover={{ scale: 1.05 }}>
                       <span className={styles.monthCardLabel}>{m.label}</span>
                       <span className={styles.monthCardValue}>{analytics?.currency} {m.total.toFixed(2)}</span>
                     </motion.div>
                   ))}
                 </div>
              </motion.div>
            )}

            {/* ── AI Insights ── */}
            {insights && insights.summary && (
              <motion.div className={`${styles.insightsCard} glass-panel`} variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }}>
                <div className={styles.cardHeader}>
                  <div className={styles.cardIcon} style={{ background: 'var(--accent-bg)', color: 'var(--accent)', border: '1px solid var(--border)' }}>
                    <Sparkles size={24} />
                  </div>
                  <div>
                    <h2>AI Spending Insights</h2>
                    {insights.computedAt && <span className={styles.timestamp}><Clock size={12} /> Last updated: {new Date(insights.computedAt).toLocaleDateString()}</span>}
                  </div>
                </div>
                <p className={styles.summary}>{insights.summary}</p>
                {insights.warnings && insights.warnings.length > 0 && (
                  <div className={styles.warnings}>
                    {insights.warnings.map((w, i) => (
                      <div key={i} className={styles.warningItem}>
                        <AlertTriangle size={18} />
                        <span>{w}</span>
                      </div>
                    ))}
                  </div>
                )}
              </motion.div>
            )}

            {/* ── Charts grid ── */}
            <div className={styles.chartsGrid}>
              {/* Category breakdown — Pie */}
              <motion.div className={`${styles.chartCard} glass-panel`} variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }}>
                <div className={styles.cardHeader}>
                  <div className={styles.cardIcon} style={{ background: 'var(--pink-bg)', color: 'var(--pink)', border: '1px solid var(--border)' }}>
                    <PieIcon size={24} />
                  </div>
                  <h2>Spending Breakdown</h2>
                </div>
                <div className={styles.chartWrap}>
                  <ResponsiveContainer width="100%" height={280}>
                    <PieChart>
                      <Pie
                        data={analytics!.categoryBreakdown}
                        dataKey="amount"
                        nameKey="category"
                        cx="50%"
                        cy="50%"
                        outerRadius={95}
                        innerRadius={55}
                        paddingAngle={5}
                        strokeWidth={0}
                        cornerRadius={4}
                      >
                        {analytics!.categoryBreakdown.map((_, i) => (
                          <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip
                        contentStyle={{
                          background: 'var(--bg-card)',
                          backdropFilter: 'blur(10px)',
                          border: '1px solid var(--border)',
                          borderRadius: '12px',
                          color: 'var(--text-primary)',
                          fontSize: '0.9rem',
                          boxShadow: 'var(--shadow-md)'
                        }}
                        formatter={(value: any) => [`${analytics!.currency} ${Number(value).toFixed(2)}`, 'Amount']}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <div className={styles.legend}>
                  {analytics!.categoryBreakdown.map((item, i) => (
                    <div key={item.category} className={styles.legendItem}>
                      <span className={styles.legendDot} style={{ background: CHART_COLORS[i % CHART_COLORS.length] }} />
                      <span className={styles.legendLabel}>{item.category}</span>
                      <span className={styles.legendValue}>{analytics!.currency} {item.amount.toFixed(0)}</span>
                    </div>
                  ))}
                </div>
              </motion.div>

              {/* Category breakdown — Bar */}
              <motion.div className={`${styles.chartCard} glass-panel`} variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }}>
                <div className={styles.cardHeader}>
                  <div className={styles.cardIcon} style={{ background: 'var(--green-bg)', color: 'var(--green)', border: '1px solid var(--border)' }}>
                    <TrendingUp size={24} />
                  </div>
                  <h2>Category Comparison</h2>
                </div>
                <div className={styles.chartWrap}>
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={analytics!.categoryBreakdown} layout="vertical" margin={{ left: 10, right: 20, top: 10, bottom: 10 }}>
                      <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" horizontal={false} />
                      <XAxis type="number" tick={{ fill: 'var(--text-muted)', fontSize: 12 }} axisLine={false} tickLine={false} />
                      <YAxis type="category" dataKey="category" width={110} tick={{ fill: 'var(--text-secondary)', fontSize: 12 }} axisLine={false} tickLine={false} />
                      <Tooltip
                        contentStyle={{ 
                          background: 'var(--bg-card)', 
                          backdropFilter: 'blur(10px)',
                          border: '1px solid var(--border)', 
                          borderRadius: '12px', 
                          color: 'var(--text-primary)', 
                          fontSize: '0.9rem',
                          boxShadow: 'var(--shadow-md)'
                        }}
                        formatter={(value: any) => [`${analytics!.currency} ${Number(value).toFixed(2)}`, 'Amount']}
                      />
                      <Bar dataKey="amount" radius={[0, 8, 8, 0]}>
                        {analytics!.categoryBreakdown.map((_, i) => (
                          <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                        ))}
                      </Bar>
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </motion.div>

              {/* Daily spending — Area */}
              {analytics!.dailySpending && analytics!.dailySpending.length > 0 && (
                <motion.div className={`${styles.chartCard} ${styles.chartCardWide} glass-panel`} variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }}>
                  <div className={styles.cardHeader}>
                    <div className={styles.cardIcon} style={{ background: 'var(--blue-bg)', color: 'var(--blue)', border: '1px solid var(--border)' }}>
                      <Calendar size={24} />
                    </div>
                    <h2>Daily Spending Trend</h2>
                  </div>
                  <div className={styles.chartWrap}>
                    <ResponsiveContainer width="100%" height={280}>
                      <AreaChart data={analytics!.dailySpending} margin={{ left: 10, right: 20, top: 20, bottom: 10 }}>
                        <defs>
                          <linearGradient id="areaGrad" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="0%" stopColor="var(--accent)" stopOpacity={0.6} />
                            <stop offset="100%" stopColor="var(--accent)" stopOpacity={0} />
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
                        <XAxis
                          dataKey="date"
                          tick={{ fill: 'var(--text-muted)', fontSize: 12 }}
                          axisLine={false}
                          tickLine={false}
                          tickFormatter={(d: string) => { const dt = new Date(d); return `${dt.getDate()}/${dt.getMonth() + 1}`; }}
                        />
                        <YAxis tick={{ fill: 'var(--text-muted)', fontSize: 12 }} axisLine={false} tickLine={false} />
                        <Tooltip
                          contentStyle={{ 
                            background: 'var(--bg-card)', 
                            backdropFilter: 'blur(10px)',
                            border: '1px solid var(--border)', 
                            borderRadius: '12px', 
                            color: 'var(--text-primary)', 
                            fontSize: '0.9rem',
                            boxShadow: 'var(--shadow-md)'
                          }}
                          formatter={(value: any) => [`${analytics!.currency} ${Number(value).toFixed(2)}`, 'Spent']}
                          labelFormatter={(d: any) => new Date(String(d)).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric'})}
                        />
                        <Area type="monotone" dataKey="amount" stroke="var(--accent)" strokeWidth={3} fill="url(#areaGrad)" activeDot={{ r: 6, fill: 'var(--accent-dark)', stroke: 'white', strokeWidth: 2 }} />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                </motion.div>
              )}
            </div>
          </motion.div>
        )}
      </main>
    </div>
  );
}
