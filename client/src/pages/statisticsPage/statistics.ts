export interface StatisticsPhoto {
  control_angajare?: boolean;
  control_periodic?: boolean;
  control_adaptare?: boolean;
  control_reluare?: boolean;
  control_supraveghere?: boolean;
  control_alte?: boolean;
  aviz_apt?: boolean;
  aviz_apt_conditionat?: boolean;
  aviz_inapt_temporar?: boolean;
  aviz_inapt?: boolean;
}

export const getControlStats = (photos: StatisticsPhoto[]) => {
  const stats = {
    Angajare: 0,
    Periodic: 0,
    Adaptare: 0,
    Reluare: 0,
    Supraveghere: 0,
    Alte: 0,
  };

  photos.forEach((photo) => {
    if (photo.control_angajare) stats.Angajare++;
    if (photo.control_periodic) stats.Periodic++;
    if (photo.control_adaptare) stats.Adaptare++;
    if (photo.control_reluare) stats.Reluare++;
    if (photo.control_supraveghere) stats.Supraveghere++;
    if (photo.control_alte) stats.Alte++;
  });

  return Object.entries(stats).map(([name, value]) => ({ name, value }));
};

export const getAvizStats = (photos: StatisticsPhoto[]) => {
  const stats = {
    APT: 0,
    'APT Conditionat': 0,
    'Inapt Temporar': 0,
    Inapt: 0,
  };

  photos.forEach((photo) => {
    if (photo.aviz_apt) stats.APT++;
    if (photo.aviz_apt_conditionat) stats['APT Conditionat']++;
    if (photo.aviz_inapt_temporar) stats['Inapt Temporar']++;
    if (photo.aviz_inapt) stats.Inapt++;
  });

  return Object.entries(stats).map(([name, value]) => ({ name, value }));
};
